package injectproxy

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func SetRecursive(node parser.Node, matchersToEnforce []*labels.Matcher) (err error) {
	switch n := node.(type) {
	case *parser.EvalStmt:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case parser.Expressions:
		for _, e := range n {
			if err := SetRecursive(e, matchersToEnforce); err != nil {
				return err
			}
		}
	case *parser.AggregateExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.BinaryExpr:
		if err := SetRecursive(n.LHS, matchersToEnforce); err != nil {
			return err
		}
		if err := SetRecursive(n.RHS, matchersToEnforce); err != nil {
			return err
		}

	case *parser.Call:
		if err := SetRecursive(n.Args, matchersToEnforce); err != nil {
			return err
		}

	case *parser.ParenExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.UnaryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.SubqueryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.NumberLiteral, *parser.StringLiteral:
	// nothing to do

	case *parser.MatrixSelector:
		// inject labelselector
		if vs, ok := n.VectorSelector.(*parser.VectorSelector); ok {
			vs.LabelMatchers = enforceLabelMatchers(vs.LabelMatchers, matchersToEnforce)
		}

	case *parser.VectorSelector:
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
