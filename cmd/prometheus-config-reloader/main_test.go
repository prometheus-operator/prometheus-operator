package main

import (
	"os"
	"testing"
)

var cases = []struct {
	in  string
	out string
}{
	{"prometheus-0", "0"},
	{"prometheus-1", "1"},
	{"prometheus-10", "10"},
	{"prometheus-10a", ""},
	{"prometheus1", "1"},
	{"pro-10-metheus", ""},
}

func TestCreateOrdinalEnvVar(t *testing.T) {
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			os.Setenv(statefulsetOrdinalFromEnvvarDefault, tt.in)
			s := createOrdinalEnvvar(statefulsetOrdinalFromEnvvarDefault)
			if os.Getenv(statefulsetOrdinalEnvvar) != tt.out {
				t.Errorf("got %v, want %s", s, tt.out)
			}
		})
	}
}

