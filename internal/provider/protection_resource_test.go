package provider

import (
	"testing"
	"time"
)

func TestFormatProtectionStartDate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "iso", in: "2026-07-20", want: "07/20/2026"},
		{name: "already api format", in: "07/20/2026", want: "07/20/2026"},
		{name: "empty stays empty", in: "", want: ""},
		{name: "invalid", in: "20-07-2026", wantErr: true},
		{name: "garbage", in: "not-a-date", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatProtectionStartDate(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("formatProtectionStartDate(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("formatProtectionStartDate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatProtectionStartTime(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "midnight 24h", in: "00:00", want: "12:00 AM"},
		{name: "2am 24h", in: "02:00", want: "2:00 AM"},
		{name: "afternoon 24h", in: "14:30", want: "2:30 PM"},
		{name: "already api format", in: "12:00 AM", want: "12:00 AM"},
		{name: "lowercase ampm", in: "2:00 am", want: "2:00 AM"},
		{name: "empty stays empty", in: "", want: ""},
		{name: "invalid", in: "25:00", wantErr: true},
		{name: "garbage", in: "noon", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatProtectionStartTime(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("formatProtectionStartTime(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("formatProtectionStartTime(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveWeekdayStartDate(t *testing.T) {
	now := time.Date(2026, time.August, 12, 20, 30, 0, 0, time.UTC) // Thursday in IST

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "same day in ist", in: "thursday", want: "2026-08-13"},
		{name: "short name", in: "fri", want: "2026-08-14"},
		{name: "mixed case and spaces", in: "  MONDAY ", want: "2026-08-17"},
		{name: "empty", in: "", want: ""},
		{name: "invalid", in: "funday", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveWeekdayStartDate(tt.in, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveWeekdayStartDate(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("resolveWeekdayStartDate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveProtectionStartDateInput(t *testing.T) {
	now := time.Date(2026, time.August, 12, 20, 30, 0, 0, time.UTC) // Thursday in IST

	t.Run("start_date takes precedence", func(t *testing.T) {
		got, err := resolveProtectionStartDateInput("2026-09-01", "monday", now)
		if err != nil {
			t.Fatalf("resolveProtectionStartDateInput err = %v", err)
		}
		if got != "2026-09-01" {
			t.Errorf("resolveProtectionStartDateInput = %q, want %q", got, "2026-09-01")
		}
	})

	t.Run("weekday derived", func(t *testing.T) {
		got, err := resolveProtectionStartDateInput("", "monday", now)
		if err != nil {
			t.Fatalf("resolveProtectionStartDateInput err = %v", err)
		}
		if got != "2026-08-17" {
			t.Errorf("resolveProtectionStartDateInput = %q, want %q", got, "2026-08-17")
		}
	})
}

func TestStringValueOrNull(t *testing.T) {
	t.Run("empty string becomes null", func(t *testing.T) {
		got := stringValueOrNull("")
		if !got.IsNull() {
			t.Fatalf("stringValueOrNull(empty) should be null")
		}
	})

	t.Run("whitespace string becomes null", func(t *testing.T) {
		got := stringValueOrNull("   ")
		if !got.IsNull() {
			t.Fatalf("stringValueOrNull(whitespace) should be null")
		}
	})

	t.Run("non-empty string becomes value", func(t *testing.T) {
		got := stringValueOrNull("south-1")
		if got.IsNull() {
			t.Fatalf("stringValueOrNull(non-empty) should not be null")
		}
		if got.ValueString() != "south-1" {
			t.Fatalf("stringValueOrNull(non-empty) = %q, want %q", got.ValueString(), "south-1")
		}
	})
}
