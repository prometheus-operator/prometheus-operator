package query

import (
	"context"
	"sort"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/improbable-eng/thanos/pkg/store/storepb"
	"github.com/improbable-eng/thanos/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

// PartialErrReporter allows to report partial errors. Partial error occurs when only part of the results are ready and
// another is not available because of the failure. We still want to return partial result, but with some notification.
// NOTE: It is required to be thread-safe.
type PartialErrReporter func(error)

// QueryableCreator returns implementation of promql.Queryable that fetches data from the proxy store API endpoints.
// If deduplication is enabled, all data retrieved from it will be deduplicated along the replicaLabel by default.
type QueryableCreator func(deduplicate bool, p PartialErrReporter) storage.Queryable

// NewQueryableCreator creates QueryableCreator.
func NewQueryableCreator(logger log.Logger, proxy storepb.StoreServer, replicaLabel string) QueryableCreator {
	return func(deduplicate bool, p PartialErrReporter) storage.Queryable {
		return &queryable{
			logger:           logger,
			replicaLabel:     replicaLabel,
			proxy:            proxy,
			deduplicate:      deduplicate,
			partialErrReport: p,
		}
	}
}

type queryable struct {
	logger           log.Logger
	replicaLabel     string
	proxy            storepb.StoreServer
	deduplicate      bool
	partialErrReport PartialErrReporter
}

// Querier returns a new storage querier against the underlying proxy store API.
func (q *queryable) Querier(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
	return newQuerier(ctx, q.logger, mint, maxt, q.replicaLabel, q.proxy, q.deduplicate, q.partialErrReport), nil
}

type querier struct {
	ctx              context.Context
	logger           log.Logger
	cancel           func()
	mint, maxt       int64
	replicaLabel     string
	proxy            storepb.StoreServer
	deduplicate      bool
	partialErrReport PartialErrReporter
}

// newQuerier creates implementation of storage.Querier that fetches data from the proxy
// store API endpoints.
func newQuerier(
	ctx context.Context,
	logger log.Logger,
	mint, maxt int64,
	replicaLabel string,
	proxy storepb.StoreServer,
	deduplicate bool,
	partialErrReport PartialErrReporter,
) *querier {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	if partialErrReport == nil {
		partialErrReport = func(error) {}
	}
	ctx, cancel := context.WithCancel(ctx)
	return &querier{
		ctx:              ctx,
		logger:           logger,
		cancel:           cancel,
		mint:             mint,
		maxt:             maxt,
		replicaLabel:     replicaLabel,
		proxy:            proxy,
		deduplicate:      deduplicate,
		partialErrReport: partialErrReport,
	}
}

func (q *querier) isDedupEnabled() bool {
	return q.deduplicate && q.replicaLabel != ""
}

type seriesServer struct {
	// This field just exist to pseudo-implement the unused methods of the interface.
	storepb.Store_SeriesServer
	ctx context.Context

	seriesSet []storepb.Series
	warnings  []string
}

func (s *seriesServer) Send(r *storepb.SeriesResponse) error {
	if r.GetWarning() != "" {
		s.warnings = append(s.warnings, r.GetWarning())
		return nil
	}

	if r.GetSeries() == nil {
		return errors.New("no seriesSet")
	}
	s.seriesSet = append(s.seriesSet, *r.GetSeries())
	return nil
}

func (s *seriesServer) Context() context.Context {
	return s.ctx
}

type resAggr int

const (
	resAggrAvg resAggr = iota
	resAggrCount
	resAggrSum
	resAggrMin
	resAggrMax
	resAggrCounter
)

// aggrsFromFunc infers aggregates of the underlying data based on the wrapping
// function of a series selection.
func aggrsFromFunc(f string) ([]storepb.Aggr, resAggr) {
	if f == "min" || strings.HasPrefix(f, "min_") {
		return []storepb.Aggr{storepb.Aggr_MIN}, resAggrMin
	}
	if f == "max" || strings.HasPrefix(f, "max_") {
		return []storepb.Aggr{storepb.Aggr_MAX}, resAggrMax
	}
	if f == "count" || strings.HasPrefix(f, "count_") {
		return []storepb.Aggr{storepb.Aggr_COUNT}, resAggrCount
	}
	if f == "sum" || strings.HasPrefix(f, "sum_") {
		return []storepb.Aggr{storepb.Aggr_SUM}, resAggrSum
	}
	if f == "increase" || f == "rate" {
		return []storepb.Aggr{storepb.Aggr_COUNTER}, resAggrCounter
	}
	// In the default case, we retrieve count and sum to compute an average.
	return []storepb.Aggr{storepb.Aggr_COUNT, storepb.Aggr_SUM}, resAggrAvg
}

func (q *querier) Select(params *storage.SelectParams, ms ...*labels.Matcher) (storage.SeriesSet, error) {
	span, ctx := tracing.StartSpan(q.ctx, "querier_select")
	defer span.Finish()

	sms, err := translateMatchers(ms...)
	if err != nil {
		return nil, errors.Wrap(err, "convert matchers")
	}

	queryAggrs, resAggr := aggrsFromFunc(params.Func)

	resp := &seriesServer{ctx: ctx}
	if err := q.proxy.Series(&storepb.SeriesRequest{
		MinTime:             q.mint,
		MaxTime:             q.maxt,
		Matchers:            sms,
		MaxResolutionWindow: params.Step / 5, // Fit at least 5 samples between steps.
		Aggregates:          queryAggrs,
	}, resp); err != nil {
		return nil, errors.Wrap(err, "proxy Series()")
	}

	for _, w := range resp.warnings {
		q.partialErrReport(errors.New(w))
	}

	if !q.isDedupEnabled() {
		// Return data without any deduplication.
		return promSeriesSet{
			mint: q.mint,
			maxt: q.maxt,
			set:  newStoreSeriesSet(resp.seriesSet),
			aggr: resAggr,
		}, nil
	}

	// TODO(fabxc): this could potentially pushed further down into the store API
	// to make true streaming possible.
	sortDedupLabels(resp.seriesSet, q.replicaLabel)

	set := promSeriesSet{
		mint: q.mint,
		maxt: q.maxt,
		set:  newStoreSeriesSet(resp.seriesSet),
		aggr: resAggr,
	}

	// The merged series set assembles all potentially-overlapping time ranges
	// of the same series into a single one. The series are ordered so that equal series
	// from different replicas are sequential. We can now deduplicate those.
	return newDedupSeriesSet(set, q.replicaLabel), nil
}

// sortDedupLabels resorts the set so that the same series with different replica
// labels are coming right after each other.
func sortDedupLabels(set []storepb.Series, replicaLabel string) {
	for _, s := range set {
		// Move the replica label to the very end.
		sort.Slice(s.Labels, func(i, j int) bool {
			if s.Labels[i].Name == replicaLabel {
				return false
			}
			if s.Labels[j].Name == replicaLabel {
				return true
			}
			return s.Labels[i].Name < s.Labels[j].Name
		})
	}
	// With the re-ordered label sets, re-sorting all series aligns the same series
	// from different replicas sequentially.
	sort.Slice(set, func(i, j int) bool {
		return storepb.CompareLabels(set[i].Labels, set[j].Labels) < 0
	})
}

func (q *querier) LabelValues(name string) ([]string, error) {
	span, ctx := tracing.StartSpan(q.ctx, "querier_label_values")
	defer span.Finish()

	resp, err := q.proxy.LabelValues(ctx, &storepb.LabelValuesRequest{Label: name})
	if err != nil {
		return nil, errors.Wrap(err, "proxy LabelValues()")
	}

	for _, w := range resp.Warnings {
		q.partialErrReport(errors.New(w))
	}

	return resp.Values, nil
}

func (q *querier) Close() error {
	q.cancel()
	return nil
}
