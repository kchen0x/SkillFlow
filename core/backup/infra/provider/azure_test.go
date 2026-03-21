//go:build !provider_select || backup_azure

package provider

import "testing"

func TestNormalizeAzureServiceURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		accountName string
		want        string
		wantErr     bool
	}{
		{
			name:        "default url from account name",
			input:       "",
			accountName: "myskillflowstorage",
			want:        "https://myskillflowstorage.blob.core.windows.net/",
		},
		{
			name:        "keeps explicit url",
			input:       "https://acct.blob.core.windows.net",
			accountName: "acct",
			want:        "https://acct.blob.core.windows.net/",
		},
		{
			name:        "adds scheme when missing",
			input:       "acct.blob.core.windows.net",
			accountName: "acct",
			want:        "https://acct.blob.core.windows.net/",
		},
		{
			name:    "rejects completely empty input",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAzureServiceURL(tt.input, tt.accountName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeAzureServiceURL(%q, %q)=%q, want %q", tt.input, tt.accountName, got, tt.want)
			}
		})
	}
}

func TestAccountNameFromAzureServiceURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "service url", input: "https://acct.blob.core.windows.net/", want: "acct"},
		{name: "host only", input: "acct.blob.core.windows.net", want: "acct"},
		{name: "empty", input: " ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accountNameFromAzureServiceURL(tt.input); got != tt.want {
				t.Fatalf("accountNameFromAzureServiceURL(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
