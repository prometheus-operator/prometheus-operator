package injectproxy

import (
	"fmt"
	"net/http"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

type routes struct {
	handler http.Handler
	label   string
}

func NewRoutes(handler http.Handler, label string) *routes {
	return &routes{
		handler: handler,
		label:   label,
	}
}

func (r *routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/api/v1/query" || req.URL.Path == "/api/v1/query_range" {
		r.query(w, req)
		return
	}
	if req.URL.Path == "/federate" {
		r.federate(w, req)
		return
	}

	http.NotFound(w, req)
}

func (r *routes) query(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	labelValue := q.Get(r.label)
	if labelValue == "" {
		http.Error(w, fmt.Sprintf("Bad request. The %q query parameter must be provided.", r.label), http.StatusBadRequest)
		return
	}
	req.URL.Query().Del(r.label)

	expr, err := promql.ParseExpr(req.FormValue("query"))
	if err != nil {
		return
	}

	err = SetRecursive(expr, []*labels.Matcher{
		{
			Name:  r.label,
			Type:  labels.MatchEqual,
			Value: labelValue,
		},
	})
	if err != nil {
		return
	}

	q.Set("query", expr.String())
	req.URL.RawQuery = q.Encode()

	r.handler.ServeHTTP(w, req)
}

func (r *routes) federate(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	labelValue := q.Get(r.label)
	if labelValue == "" {
		http.Error(w, fmt.Sprintf("Bad request. The %q query parameter must be provided.", r.label), http.StatusBadRequest)
		return
	}
	req.URL.Query().Del(r.label)

	matcher := &labels.Matcher{
		Name:  r.label,
		Type:  labels.MatchEqual,
		Value: labelValue,
	}

	q.Set("match[]", "{"+matcher.String()+"}")
	req.URL.RawQuery = q.Encode()

	r.handler.ServeHTTP(w, req)
}
