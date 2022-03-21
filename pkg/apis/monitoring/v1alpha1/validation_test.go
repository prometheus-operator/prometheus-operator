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
	"testing"
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
			name:      "Test invalid months returns error",
			in:        MonthRange("januarE"),
			expectErr: true,
		},
		{
			name:      "Test invalid months in range returns error",
			in:        MonthRange("january:Merch"),
			expectErr: true,
		},
		{
			name:      "Test invalid range - end before start returns error",
			in:        MonthRange("march:january"),
			expectErr: true,
		},
		{
			name: "Test happy path",
			in:   MonthRange("january"),
			expectResult: &ParsedRange{
				Start: 1,
				End:   1,
			},
		},
		{
			name: "Test happy path range",
			in:   MonthRange("january:march"),
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
