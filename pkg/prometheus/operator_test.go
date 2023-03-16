package prometheus

import "testing"

func TestStatefulSetKeyToPrometheusKey(t *testing.T) {
	cases := []struct {
		input         string
		expectedKey   string
		expectedMatch bool
	}{
		{
			input:         "namespace/prometheus-test",
			expectedKey:   "namespace/test",
			expectedMatch: true,
		},
		{
			input:         "namespace/prometheus-test-shard-1",
			expectedKey:   "namespace/test",
			expectedMatch: true,
		},
		{
			input:         "allns-z-thanosrulercreatedeletecluster-qcwdmj-0/thanos-ruler-test",
			expectedKey:   "",
			expectedMatch: false,
		},
	}

	for _, c := range cases {
		match, key := StatefulSetKeyToPrometheusKey(c.input)
		if c.expectedKey != key {
			t.Fatalf("Expected prometheus key %q got %q", c.expectedKey, key)
		}
		if c.expectedMatch != match {
			notExp := ""
			if !c.expectedMatch {
				notExp = "not "
			}
			t.Fatalf("Expected input %sto be matching a prometheus key, but did not", notExp)
		}
	}
}


func TestKeyToStatefulSetKey(t *testing.T) {
	cases := []struct {
		name     string
		shard    int
		expected string
	}{
		{
			name:     "namespace/test",
			shard:    0,
			expected: "namespace/prometheus-test",
		},
		{
			name:     "namespace/test",
			shard:    1,
			expected: "namespace/prometheus-test-shard-1",
		},
	}

	for _, c := range cases {
		got := KeyToStatefulSetKey(c.name, c.shard)
		if c.expected != got {
			t.Fatalf("Expected key %q got %q", c.expected, got)
		}
	}
}