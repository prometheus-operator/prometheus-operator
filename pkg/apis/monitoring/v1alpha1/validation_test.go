// Copyright 2021 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"reflect"
	"strings"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestTimeRange_Parse(t *testing.T) {
	testCases := []struct {
		name         string
		in           TimeRange
		expectErr    bool
		expectResult *ParsedRange
	}{
		{
			name: "Test invalid time string produces error",
			in: TimeRange{
				StartTime: "16:00",
				EndTime:   "25:00",
			},
			expectErr: true,
		},
		{
			name: "Test invalid negative string produces error",
			in: TimeRange{
				StartTime: "-16:00",
				EndTime:   "24:00",
			},
			expectErr: true,
		},
		{
			name: "Test end time earlier than start time is invalid",
			in: TimeRange{
				StartTime: "16:00",
				EndTime:   "14:00",
			},
			expectErr: true,
		},
		{
			name: "Test happy path",
			in: TimeRange{
				StartTime: "12:00",
				EndTime:   "24:00",
			},
			expectResult: &ParsedRange{
				Start: 720,
				End:   1440,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Parse()
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if !reflect.DeepEqual(got, tc.expectResult) {
				t.Fatalf("wanted %v, but got %v", tc.expectResult, got)
			}
		})
	}
}

func TestMonthRange_Parse(t *testing.T) {
	testCases := []struct {
		name         string
		in           MonthRange
		expectErr    bool
		expectResult *ParsedRange
	}{
		{
			name:      "Test invalid range - more than two months returns an error",
			in:        MonthRange("january:march:december"),
			expectErr: true,
		},
		{
			name:      "Test invalid named months returns error",
			in:        MonthRange("januarE"),
			expectErr: true,
		},
		{
			name:      "Test invalid numerical months returns error",
			in:        MonthRange("13"),
			expectErr: true,
		},
		{
			name:      "Test invalid named months in range returns error",
			in:        MonthRange("january:Merch"),
			expectErr: true,
		},
		{
			name:      "Test invalid numerical months in range returns error",
			in:        MonthRange("1:13"),
			expectErr: true,
		},
		{
			name:      "Test invalid named range - end before start returns error",
			in:        MonthRange("march:january"),
			expectErr: true,
		},
		{
			name:      "Test invalid numerical range - end before start returns error",
			in:        MonthRange("3:1"),
			expectErr: true,
		},
		{
			name: "Test happy named path",
			in:   MonthRange("january"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   1,
			},
		},
		{
			name: "Test happy one digit numerical path",
			in:   MonthRange("1"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   1,
			},
		},
		{
			name: "Test happy two digits numerical path",
			in:   MonthRange("12"),
			expectResult: &ParsedRange{
				Start: 12,
				End:   12,
			},
		},
		{
			name: "Test happy named path range",
			in:   MonthRange("january:march"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   3,
			},
		},
		{
			name: "Test happy numerical path range",
			in:   MonthRange("1:12"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   12,
			},
		},
		{
			name: "Test happy mixed path range",
			in:   MonthRange("1:march"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Parse()
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if !reflect.DeepEqual(got, tc.expectResult) {
				t.Fatalf("wanted %v, but got %v", tc.expectResult, got)
			}
		})
	}
}

func TestWeekdayRange_Parse(t *testing.T) {
	testCases := []struct {
		name         string
		in           WeekdayRange
		expectErr    bool
		expectResult *ParsedRange
	}{
		{
			name:      "Test invalid range - more than two days returns an error",
			in:        WeekdayRange("monday:wednesday:friday"),
			expectErr: true,
		},
		{
			name:      "Test invalid day returns error",
			in:        WeekdayRange("onday"),
			expectErr: true,
		},
		{
			name:      "Test invalid days in range returns error",
			in:        WeekdayRange("monday:friyay"),
			expectErr: true,
		},
		{
			name:      "Test invalid range - end before start returns error",
			in:        WeekdayRange("friday:monday"),
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   WeekdayRange("monday"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   1,
			},
		},
		{
			name: "Test happy path range",
			in:   WeekdayRange("monday:wednesday"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Parse()
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if !reflect.DeepEqual(got, tc.expectResult) {
				t.Fatalf("wanted %v, but got %v", tc.expectResult, got)
			}
		})
	}
}

func TestDayOfMonthRange_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		in        DayOfMonthRange
		expectErr bool
	}{
		{
			name: "Test zero value returns error",
			in: DayOfMonthRange{
				Start: 0,
				End:   0,
			},
			expectErr: true,
		},
		{
			name: "Test out of range returns error",
			in: DayOfMonthRange{
				Start: -50,
				End:   -20,
			},
			expectErr: true,
		},
		{
			name: "Test out of range returns error",
			in: DayOfMonthRange{
				Start: 20,
				End:   50,
			},
			expectErr: true,
		},
		{
			name: "Test invalid input - negative start day with positive end day",
			in: DayOfMonthRange{
				Start: -20,
				End:   5,
			},
			expectErr: true,
		},
		{
			name: "Test invalid range - end before start returns error",
			in: DayOfMonthRange{
				Start: 10,
				End:   -25,
			},
			expectErr: true,
		},
		{
			name: "Test happy path",
			in: DayOfMonthRange{
				Start: 1,
				End:   31,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
		})
	}
}

func TestYearRange_Parse(t *testing.T) {
	testCases := []struct {
		name         string
		in           YearRange
		expectErr    bool
		expectResult *ParsedRange
	}{
		{
			name:      "Test invalid range - more than two years returns an error",
			in:        YearRange("2019:2029:2039"),
			expectErr: true,
		},
		{
			name:      "Test invalid range - end before start returns error",
			in:        YearRange("2020:2010"),
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   YearRange("2030"),
			expectResult: &ParsedRange{
				Start: 2030,
				End:   2030,
			},
		},
		{
			name: "Test happy path range",
			in:   YearRange("2030:2050"),
			expectResult: &ParsedRange{
				Start: 2030,
				End:   2050,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Parse()
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if !reflect.DeepEqual(got, tc.expectResult) {
				t.Fatalf("wanted %v, but got %v", tc.expectResult, got)
			}
		})
	}
}

func TestHTTPClientConfigValidate(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   *HTTPConfig
		fail bool
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   &HTTPConfig{},
		},
		{
			name: "duplicate basic-auth and auth",
			in: &HTTPConfig{
				Authorization: &monitoringv1.SafeAuthorization{
					Credentials: &v1.SecretKeySelector{},
				},
				BasicAuth: &monitoringv1.BasicAuth{},
			},
			fail: true,
		},
		{
			name: "duplicate basic-auth and oauth2",
			in: &HTTPConfig{
				OAuth2:    &monitoringv1.OAuth2{},
				BasicAuth: &monitoringv1.BasicAuth{},
			},
			fail: true,
		},
		{
			name: "invalid Proxy URL",
			in: &HTTPConfig{
				ProxyConfig: monitoringv1.ProxyConfig{
					ProxyURL: ptr.To("://example.com"),
				},
			},
			fail: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.in.Validate()
			if tc.fail {
				if err == nil {
					t.Fatal("expecting error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("expecting no error, got %q", err)
			}
		})
	}

}

func TestOpsGenieConfigResponder_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		in          *OpsGenieConfigResponder
		expectedErr bool
	}{
		{
			name: "Test nil ID, Name and Username",
			in: &OpsGenieConfigResponder{
				Type: "user",
			},
			expectedErr: true,
		},
		{
			name: "Test invalid template string type",
			in: &OpsGenieConfigResponder{
				Name: "responder",
				Type: "{{.GroupLabels",
			},
			expectedErr: true,
		},
		{
			name: "Test valid template string type",
			in: &OpsGenieConfigResponder{
				Name: "responder",
				Type: "{{.GroupLabels}}",
			},
			expectedErr: false,
		},
		{
			name: "Test invalid type",
			in: &OpsGenieConfigResponder{
				Name: "responder",
				Type: "username",
			},
			expectedErr: true,
		},
		{
			name: "Test valid type",
			in: &OpsGenieConfigResponder{
				Name: "responder",
				Type: "user",
			},
			expectedErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected err but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
		})
	}
}

func TestMatchers(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []*Matcher
		err   string
	}{
		{
			input: `{}`,
			want:  make([]*Matcher, 0),
		},
		{
			input: `,`,
			err:   "bad matcher format: ",
		},
		{
			input: `{,}`,
			err:   "bad matcher format: ",
		},
		{
			input: `{foo='}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "'")
				return append(ms, m)
			}(),
		},
		{
			input: "{foo=`}",
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "`")
				return append(ms, m)
			}(),
		},
		{
			input: "{foo=\\\"}",
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "\"")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=bar}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo="bar"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=~bar.*}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchRegexp, "foo", "bar.*")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=~"bar.*"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchRegexp, "foo", "bar.*")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo!=bar}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchNotEqual, "foo", "bar")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo!="bar"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchNotEqual, "foo", "bar")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo!~bar.*}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchNotRegexp, "foo", "bar.*")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo!~"bar.*"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchNotRegexp, "foo", "bar.*")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo="bar", baz!="quux"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotEqual, "baz", "quux")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{foo="bar", baz!~"quux.*"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotRegexp, "baz", "quux.*")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{foo="bar",baz!~".*quux", derp="wat"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotRegexp, "baz", ".*quux")
				m3, _ := NewMatcher(MatchEqual, "derp", "wat")
				return append(ms, m, m2, m3)
			}(),
		},
		{
			input: `{foo="bar", baz!="quux", derp="wat"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotEqual, "baz", "quux")
				m3, _ := NewMatcher(MatchEqual, "derp", "wat")
				return append(ms, m, m2, m3)
			}(),
		},
		{
			input: `{foo="bar", baz!~".*quux.*", derp="wat"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotRegexp, "baz", ".*quux.*")
				m3, _ := NewMatcher(MatchEqual, "derp", "wat")
				return append(ms, m, m2, m3)
			}(),
		},
		{
			input: `{foo="bar", instance=~"some-api.*"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchRegexp, "instance", "some-api.*")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{foo=""}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo="bar,quux", job="job1"}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar,quux")
				m2, _ := NewMatcher(MatchEqual, "job", "job1")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{foo = "bar", dings != "bums", }`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotEqual, "dings", "bums")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `foo=bar,dings!=bums`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar")
				m2, _ := NewMatcher(MatchNotEqual, "dings", "bums")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{quote="She said: \"Hi, ladies! That's gender-neutral…\""}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "quote", `She said: "Hi, ladies! That's gender-neutral…"`)
				return append(ms, m)
			}(),
		},
		{
			input: `statuscode=~"5.."`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchRegexp, "statuscode", "5..")
				return append(ms, m)
			}(),
		},
		{
			input: `tricky=~~~`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchRegexp, "tricky", "~~")
				return append(ms, m)
			}(),
		},
		{
			input: `trickier==\\=\=\"`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "trickier", `=\=\="`)
				return append(ms, m)
			}(),
		},
		{
			input: `contains_quote != "\"" , contains_comma !~ "foo,bar" , `,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchNotEqual, "contains_quote", `"`)
				m2, _ := NewMatcher(MatchNotRegexp, "contains_comma", "foo,bar")
				return append(ms, m, m2)
			}(),
		},
		{
			input: `{foo=bar}}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar}")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=bar}},}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar}}")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=,bar=}}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m1, _ := NewMatcher(MatchEqual, "foo", "")
				m2, _ := NewMatcher(MatchEqual, "bar", "}")
				return append(ms, m1, m2)
			}(),
		},
		{
			input: `{foo=bar\t}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar\\t")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=bar\n}`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "bar\n")
				return append(ms, m)
			}(),
		},
		{
			input: `job=`,
			want: func() []*Matcher {
				m, _ := NewMatcher(MatchEqual, "job", "")
				return []*Matcher{m}
			}(),
		},
		{
			input: `job="value`,
			err:   `matcher value contains unescaped double quote: "value`,
		},
		{
			input: `job=value"`,
			err:   `matcher value contains unescaped double quote: value"`,
		},
		{
			input: `trickier==\\=\=\""`,
			err:   `matcher value contains unescaped double quote: =\\=\=\""`,
		},
		{
			input: `contains_unescaped_quote = foo"bar`,
			err:   `matcher value contains unescaped double quote: foo"bar`,
		},
		{
			input: `{invalid-name = "valid label"}`,
			err:   `bad matcher format: invalid-name = "valid label"`,
		},
		{
			input: `{foo=~"invalid[regexp"}`,
			err:   "error parsing regexp: missing closing ]: `[regexp)$`",
		},
		// Double escaped strings.
		{
			input: `"{foo=\"bar"}`,
			err:   `bad matcher format: "{foo=\"bar"`,
		},
		{
			input: `"foo=\"bar"`,
			err:   `bad matcher format: "foo=\"bar"`,
		},
		{
			input: `"foo=\"bar\""`,
			err:   `bad matcher format: "foo=\"bar\""`,
		},
		{
			input: `"foo=\"bar\"`,
			err:   `bad matcher format: "foo=\"bar\"`,
		},
		{
			input: `"{foo=\"bar\"}"`,
			err:   `bad matcher format: "{foo=\"bar\"}"`,
		},
		{
			input: `"foo="bar""`,
			err:   `bad matcher format: "foo="bar""`,
		},
		{
			input: `{{foo=`,
			err:   `bad matcher format: {foo=`,
		},
		{
			input: `{foo=`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "")
				return append(ms, m)
			}(),
		},
		{
			input: `{foo=}b`,
			want: func() []*Matcher {
				ms := []*Matcher{}
				m, _ := NewMatcher(MatchEqual, "foo", "}b")
				return append(ms, m)
			}(),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if tc.input == "" {
				t.Skip("empty matchers are implicitly ignored, skipping")
			}
			got, err := ParseMatchers(tc.input)
			if err != nil && tc.err == "" {
				t.Fatalf("got error where none expected: %v", err)
			}
			if err == nil && tc.err != "" {
				t.Fatalf("expected error but got none: %v", tc.err)
			}
			if err != nil && err.Error() != tc.err {
				t.Fatalf("error not equal:\ngot  %v\nwant %v", err, tc.err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("labels not equal:\ngot  %v\nwant %v", got, tc.want)
			}
		})
	}
}

// ParseMatchers parses a comma-separated list of Matchers. A leading '{' and/or
// a trailing '}' is optional and will be trimmed before further
// parsing. Individual Matchers are separated by commas outside of quoted parts
// of the input string. Those commas may be surrounded by whitespace. Parts of the
// string inside unescaped double quotes ('"…"') are considered quoted (and
// commas don't act as separators there). If double quotes are escaped with a
// single backslash ('\"'), they are ignored for the purpose of identifying
// quoted parts of the input string. If the input string, after trimming the
// optional trailing '}', ends with a comma, followed by optional whitespace,
// this comma and whitespace will be trimmed.
//
// Examples for valid input strings:
//
//	{foo = "bar", dings != "bums", }
//	foo=bar,dings!=bums
//	foo=bar, dings!=bums
//	{quote="She said: \"Hi, ladies! That's gender-neutral…\""}
//	statuscode=~"5.."
//
// See ParseMatcher for details on how an individual Matcher is parsed.
func ParseMatchers(s string) ([]*Matcher, error) {
	matchers := []*Matcher{}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	var (
		insideQuotes bool
		escaped      bool
		token        strings.Builder
		tokens       []string
	)
	for _, r := range s {
		switch r {
		case ',':
			if !insideQuotes {
				tokens = append(tokens, token.String())
				token.Reset()
				continue
			}
		case '"':
			if !escaped {
				insideQuotes = !insideQuotes
			} else {
				escaped = false
			}
		case '\\':
			escaped = !escaped
		default:
			escaped = false
		}
		token.WriteRune(r)
	}
	if s := strings.TrimSpace(token.String()); s != "" {
		tokens = append(tokens, s)
	}
	for _, token := range tokens {
		m, err := ParseMatcher(token)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}

	return matchers, nil
}
