package k8s

import "testing"

func TestLabelSelector(t *testing.T) {
	tests := []struct {
		f    func(l *LabelSelector)
		want string
	}{
		{
			f: func(l *LabelSelector) {
				l.Eq("component", "frontend")
			},
			want: "component=frontend",
		},
		{
			f: func(l *LabelSelector) {
				l.Eq("kubernetes.io/role", "master")
			},
			want: "kubernetes.io/role=master",
		},
		{
			f: func(l *LabelSelector) {
				l.In("type", "prod", "staging")
				l.Eq("component", "frontend")
			},
			want: "type in (prod, staging),component=frontend",
		},
		{
			f: func(l *LabelSelector) {
				l.NotIn("type", "prod", "staging")
				l.NotEq("component", "frontend")
			},
			want: "type notin (prod, staging),component!=frontend",
		},
		{
			f: func(l *LabelSelector) {
				l.Eq("foo", "I am not a valid label value")
			},
			want: "",
		},
	}

	for i, test := range tests {
		l := new(LabelSelector)
		test.f(l)
		got := l.String()
		if test.want != got {
			t.Errorf("case %d: want=%q, got=%q", i, test.want, got)
		}
	}
}
