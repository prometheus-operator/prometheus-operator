package injectproxy

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

func SetRecursive(node promql.Node, matchersToEnforce []*labels.Matcher) (err error) {
	switch n := node.(type) {
	case *promql.EvalStmt:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case promql.Expressions:
		for _, e := range n {
			if err := SetRecursive(e, matchersToEnforce); err != nil {
				return err
			}
		}
	case *promql.AggregateExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *promql.BinaryExpr:
		if err := SetRecursive(n.LHS, matchersToEnforce); err != nil {
			return err
		}
		if err := SetRecursive(n.RHS, matchersToEnforce); err != nil {
			return err
		}

	case *promql.Call:
		if err := SetRecursive(n.Args, matchersToEnforce); err != nil {
			return err
		}

	case *promql.ParenExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *promql.UnaryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *promql.SubqueryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *promql.NumberLiteral, *promql.StringLiteral:
	// nothing to do

	case *promql.MatrixSelector:
		// inject labelselector
		n.LabelMatchers = enforceLabelMatchers(n.LabelMatchers, matchersToEnforce)

	case *promql.VectorSelector:
		// inject labelselector
		n.LabelMatchers = enforceLabelMatchers(n.LabelMatchers, matchersToEnforce)

	default:
		panic(fmt.Errorf("promql.Walk: unhandled node type %T", node))
	}

	return err
}

func enforceLabelMatchers(matchers []*labels.Matcher, matchersToEnforce []*labels.Matcher) []*labels.Matcher {
	res := []*labels.Matcher{}
	for _, m := range matchersToEnforce {
		res = enforceLabelMatcher(matchers, m)
	}

	return res
}

func enforceLabelMatcher(matchers []*labels.Matcher, enforcedMatcher *labels.Matcher) []*labels.Matcher {
	res := []*labels.Matcher{}
	for _, m := range matchers {
		if m.Name == enforcedMatcher.Name {
			continue
		}
		res = append(res, m)
	}

	return append(res, enforcedMatcher)
}
