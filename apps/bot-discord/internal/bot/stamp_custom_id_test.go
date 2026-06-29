package bot

import (
	"testing"
)

func TestParseStampCustomID_Valid(t *testing.T) {
	cases := []struct {
		name        string
		customID    string
		wantRoundID string
		wantJumpID  string
		wantStampID string
	}{
		{"simple ids", "stamp:round-1:jump-7:stamp-hype", "round-1", "jump-7", "stamp-hype"},
		{"longer stamp id", "stamp:abc:def:very-long-stamp-id-123", "abc", "def", "very-long-stamp-id-123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotRoundID, gotJumpID, gotStampID, ok := ParseStampCustomID(tc.customID)
			if !ok {
				t.Fatalf("ParseStampCustomID: got !ok, want ok for %q", tc.customID)
			}
			if gotRoundID != tc.wantRoundID {
				t.Errorf("roundID: got %q, want %q", gotRoundID, tc.wantRoundID)
			}
			if gotJumpID != tc.wantJumpID {
				t.Errorf("jumpID: got %q, want %q", gotJumpID, tc.wantJumpID)
			}
			if gotStampID != tc.wantStampID {
				t.Errorf("stampID: got %q, want %q", gotStampID, tc.wantStampID)
			}
		})
	}
}

func TestParseStampCustomID_Invalid(t *testing.T) {
	cases := []struct {
		name     string
		customID string
	}{
		{"empty", ""},
		{"missing prefix", "round-1:jump-7:stamp-hype"},
		{"wrong prefix", "react:round-1:jump-7:stamp-hype"},
		{"too few parts", "stamp:round-1:jump-7"},
		{"too many parts", "stamp:round-1:jump-7:stamp-hype:extra"},
		{"empty round", "stamp::jump-7:stamp-hype"},
		{"empty jump", "stamp:round-1::stamp-hype"},
		{"empty stamp", "stamp:round-1:jump-7:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, ok := ParseStampCustomID(tc.customID)
			if ok {
				t.Errorf("ParseStampCustomID(%q): got ok, want !ok", tc.customID)
			}
		})
	}
}
