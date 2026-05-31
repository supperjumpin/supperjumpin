package game

import "testing"

func TestValidScore_AcceptsBoundaryValues(t *testing.T) {
	cases := []struct {
		score int
		want  bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{-1, false},
		{10, false},
	}
	for _, tc := range cases {
		got := validScore(tc.score)
		if got != tc.want {
			t.Errorf("validScore(%d) = %v; want %v", tc.score, got, tc.want)
		}
	}
}

func TestValidScore_AllFourScoresMustBeValid(t *testing.T) {
	if !validScore(1) {
		t.Error("validScore(1) should be true (minimum valid)")
	}
	if !validScore(4) {
		t.Error("validScore(4) should be true (maximum valid)")
	}
	if !validScore(2) {
		t.Error("validScore(2) should be true")
	}
	if !validScore(3) {
		t.Error("validScore(3) should be true")
	}
}
