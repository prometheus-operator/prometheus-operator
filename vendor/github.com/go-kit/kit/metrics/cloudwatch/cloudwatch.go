package cloudwatch

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
	"github.com/go-kit/kit/metrics/internal/lv"
	"strconv"
)

const (
	maxConcurrentRequests = 20
)

type Percentiles []struct {
	s string
	f float64
}

// CloudWatch receives metrics observations and forwards them to CloudWatch.
// Create a CloudWatch object, use it to create metrics, and pass those metrics as
// dependencies to the components that will use them.
//
// To regularly report metrics to CloudWatch, use the WriteLoop helper method.
type CloudWatch struct {
	mtx                   sync.RWMutex
	sem                   chan struct{}
	namespace             string
	svc                   cloudwatchiface.CloudWatchAPI
	counters              *lv.Space
	gauges                *lv.Space
	histograms            *lv.Space
	percentiles           []float64 // percentiles to track
	logger                log.Logger
	numConcurrentRequests int
}

type option func(*CloudWatch)

func (s *CloudWatch) apply(opt option) {
	if opt != nil {
		opt(s)
	}
}

func WithLogger(logger log.Logger) option {
	return func(c *CloudWatch) {
		c.logger = logger
	}
}

// WithPercentiles registers the percentiles to track, overriding the
// existing/default values.
// Reason is that Cloudwatch makes you pay per metric, so you can save half the money
// by only using 2 metrics instead of the default 4.
func WithPercentiles(percentiles ...float64) option {
	return func(c *CloudWatch) {
		c.percentiles = make([]float64, 0, len(percentiles))
		for _, p := range percentiles {
			if p < 0 || p > 1 {
				continue // illegal entry; ignore
			}
			c.percentiles = append(c.percentiles, p)
		}
	}
}

func WithConcurrentRequests(n int) option {
	return func(c *CloudWatch) {
		if n > maxConcurrentRequests {
			n = maxConcurrentRequests
		}
		c.numConcurrentRequests = n
	}
}

// New returns a CloudWatch object that may be used to create metrics.
// Namespace is applied to all created metrics and maps to the CloudWatch namespace.
// Callers must ensure that regular calls to Send are performed, either
// manually or with one of the helper methods.
func New(namespace string, svc cloudwatchiface.CloudWatchAPI, options ...option) *CloudWatch {
	cw := &CloudWatch{
		sem:                   nil, // set below
		namespace:             namespace,
		svc:                   svc,
		counters:              lv.NewSpace(),
		gauges:                lv.NewSpace(),
		histograms:            lv.NewSpace(),
		numConcurrentRequests: 10,
		logger:                log.NewLogfmtLogger(os.Stderr),
		percentiles:           []float64{0.50, 0.90, 0.95, 0.99},
	}

	for _, optFunc := range options {
		optFunc(cw)
	}

	cw.sem = make(chan struct{}, cw.numConcurrentRequests)

	return cw
}

// NewCounter returns a counter. Observations are aggregated and emitted once
// per write invocation.
func (cw *CloudWatch) NewCounter(name string) metrics.Counter {
	return &Counter{
		name: name,
		obs:  cw.counters.Observe,
	}
}

// NewGauge returns an gauge.
func (cw *CloudWatch) NewGauge(name string) metrics.Gauge {
	return &Gauge{
		name: name,
		obs:  cw.gauges.Observe,
		add:  cw.gauges.Add,
	}
}

// NewHistogram returns a histogram.
func (cw *CloudWatch) NewHistogram(name string) metrics.Histogram {
	return &Histogram{
		name: name,
		obs:  cw.histograms.Observe,
	}
}

// WriteLoop is a helper method that invokes Send every time the passed
// channel fires. This method blocks until the channel is closed, so clients
// probably want to run it in its own goroutine. For typical usage, create a
// time.Ticker and pass its C channel to this method.
func (cw *CloudWatch) WriteLoop(c <-chan time.Time) {
	for range c {
		if err := cw.Send(); err != nil {
			cw.logger.Log("during", "Send", "err", err)
		}
	}
}

// Send will fire an API request to CloudWatch with the latest stats for
// all metrics. It is preferred that the WriteLoop method is used.
func (cw *CloudWatch) Send() error {
	cw.mtx.RLock()
	defer cw.mtx.RUnlock()
	now := time.Now()

	var datums []*cloudwatch.MetricDatum

	cw.counters.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
		value := sum(values)
		datums = append(datums, &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: makeDimensions(lvs...),
			Value:      aws.Float64(value),
			Timestamp:  aws.Time(now),
		})
		return true
	})

	cw.gauges.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
		value := last(values)
		datums = append(datums, &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: makeDimensions(lvs...),
			Value:      aws.Float64(value),
			Timestamp:  aws.Time(now),
		})
		return true
	})

	// format a [0,1]-float value to a percentile value, with minimum nr of decimals
	// 0.90 -> "90"
	// 0.95 -> "95"
	// 0.999 -> "99.9"
	formatPerc := func(p float64) string {
		return strconv.FormatFloat(p*100, 'f', -1, 64)
	}

	cw.histograms.Reset().Walk(func(name string, lvs lv.LabelValues, values []float64) bool {
		histogram := generic.NewHistogram(name, 50)

		for _, v := range values {
			histogram.Observe(v)
		}

		for _, perc := range cw.percentiles {
			value := histogram.Quantile(perc)
			datums = append(datums, &cloudwatch.MetricDatum{
				MetricName: aws.String(fmt.Sprintf("%s_%s", name, formatPerc(perc))),
				Dimensions: makeDimensions(lvs...),
				Value:      aws.Float64(value),
				Timestamp:  aws.Time(now),
			})
		}
		return true
	})

	var batches [][]*cloudwatch.MetricDatum
	for len(datums) > 0 {
		var batch []*cloudwatch.MetricDatum
		lim := min(len(datums), maxConcurrentRequests)
		batch, datums = datums[:lim], datums[lim:]
		batches = append(batches, batch)
	}

	var errors = make(chan error, len(batches))
	for _, batch := range batches {
		go func(batch []*cloudwatch.MetricDatum) {
			cw.sem <- struct{}{}
			defer func() {
				<-cw.sem
			}()
			_, err := cw.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace:  aws.String(cw.namespace),
				MetricData: batch,
			})
			errors <- err
		}(batch)
	}
	var firstErr error
	for i := 0; i < cap(errors); i++ {
		if err := <-errors; err != nil && firstErr != nil {
			firstErr = err
		}
	}

	return firstErr
}

func sum(a []float64) float64 {
	var v float64
	for _, f := range a {
		v += f
	}
	return v
}

func last(a []float64) float64 {
	return a[len(a)-1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type observeFunc func(name string, lvs lv.LabelValues, value float64)

// Counter is a counter. Observations are forwarded to a node
// object, and aggregated (summed) per timeseries.
type Counter struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
}

// With implements metrics.Counter.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	return &Counter{
		name: c.name,
		lvs:  c.lvs.With(labelValues...),
		obs:  c.obs,
	}
}

// Add implements metrics.Counter.
func (c *Counter) Add(delta float64) {
	c.obs(c.name, c.lvs, delta)
}

// Gauge is a gauge. Observations are forwarded to a node
// object, and aggregated (the last observation selected) per timeseries.
type Gauge struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
	add  observeFunc
}

// With implements metrics.Gauge.
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	return &Gauge{
		name: g.name,
		lvs:  g.lvs.With(labelValues...),
		obs:  g.obs,
		add:  g.add,
	}
}

// Set implements metrics.Gauge.
func (g *Gauge) Set(value float64) {
	g.obs(g.name, g.lvs, value)
}

// Add implements metrics.Gauge.
func (g *Gauge) Add(delta float64) {
	g.add(g.name, g.lvs, delta)
}

// Histogram is an Influx histrogram. Observations are aggregated into a
// generic.Histogram and emitted as per-quantile gauges to the Influx server.
type Histogram struct {
	name string
	lvs  lv.LabelValues
	obs  observeFunc
}

// With implements metrics.Histogram.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	return &Histogram{
		name: h.name,
		lvs:  h.lvs.With(labelValues...),
		obs:  h.obs,
	}
}

// Observe implements metrics.Histogram.
func (h *Histogram) Observe(value float64) {
	h.obs(h.name, h.lvs, value)
}

func makeDimensions(labelValues ...string) []*cloudwatch.Dimension {
	dimensions := make([]*cloudwatch.Dimension, len(labelValues)/2)
	for i, j := 0, 0; i < len(labelValues); i, j = i+2, j+1 {
		dimensions[j] = &cloudwatch.Dimension{
			Name:  aws.String(labelValues[i]),
			Value: aws.String(labelValues[i+1]),
		}
	}
	return dimensions
}
