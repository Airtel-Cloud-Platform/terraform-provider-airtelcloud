package provider

import "testing"

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
