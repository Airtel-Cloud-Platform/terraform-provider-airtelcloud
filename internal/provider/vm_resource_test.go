package provider

import (
	"strings"
	"testing"
)

func TestValidateVMLabels(t *testing.T) {
	tests := []struct {
		name    string
		labels  []string
		wantErr string
	}{
		{
			name:   "valid plain labels",
			labels: []string{"example", "web-server"},
		},
		{
			name:   "accepts three character label",
			labels: []string{"web", "prod"},
		},
		{
			name:    "rejects more than five labels",
			labels:  []string{"label1", "label2", "label3", "label4", "label5", "label6"},
			wantErr: "at most 5 labels",
		},
		{
			name:    "rejects label shorter than three characters",
			labels:  []string{"ab"},
			wantErr: "at least 3 characters",
		},
		{
			name:    "rejects label longer than fifteen characters",
			labels:  []string{"this-label-is-way-too-long"},
			wantErr: "at most 15 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVMLabelValues(tt.labels)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateVMLabelValues() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateVMLabelValues() error = nil, want substring %q", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateVMLabelValues() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}
