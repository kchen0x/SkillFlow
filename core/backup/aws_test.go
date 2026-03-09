//go:build !provider_select || backup_aws

package backup

import "testing"

func TestNormalizeAWSRegion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trim whitespace", input: "  us-east-1  ", want: "us-east-1"},
		{name: "keeps region", input: "ap-southeast-1", want: "ap-southeast-1"},
		{name: "empty stays empty", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAWSRegion(tt.input); got != tt.want {
				t.Fatalf("normalizeAWSRegion(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
