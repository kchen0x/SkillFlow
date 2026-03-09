//go:build !provider_select || backup_tencent

package backup

import "testing"

func TestNormalizeTencentBucketName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain bucket", input: "mybucket-1250000000", want: "mybucket-1250000000"},
		{name: "trimmed bucket", input: "  mybucket-1250000000  ", want: "mybucket-1250000000"},
		{name: "full host", input: "mybucket-1250000000.cos.ap-guangzhou.myqcloud.com", want: "mybucket-1250000000"},
		{name: "full url", input: "https://mybucket-1250000000.cos.ap-guangzhou.myqcloud.com/skillflow/", want: "mybucket-1250000000"},
		{name: "endpoint only", input: "cos.ap-guangzhou.myqcloud.com", wantErr: true},
		{name: "uppercase invalid", input: "MyBucket-1250000000", wantErr: true},
		{name: "empty invalid", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTencentBucketName(tt.input)
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
				t.Fatalf("normalizeTencentBucketName(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeTencentEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain endpoint", input: "cos.ap-guangzhou.myqcloud.com", want: "cos.ap-guangzhou.myqcloud.com"},
		{name: "full url", input: "https://cos.ap-guangzhou.myqcloud.com", want: "cos.ap-guangzhou.myqcloud.com"},
		{name: "bucket host", input: "https://mybucket-1250000000.cos.ap-guangzhou.myqcloud.com/", want: "cos.ap-guangzhou.myqcloud.com"},
		{name: "empty invalid", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTencentEndpoint(tt.input)
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
				t.Fatalf("normalizeTencentEndpoint(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildTencentBucketURLWithBucketHostEndpoint(t *testing.T) {
	u, err := buildTencentBucketURL("shinerio-1258556983", "shinerio-1258556983.cos.ap-guangzhou.myqcloud.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := u.String(); got != "https://shinerio-1258556983.cos.ap-guangzhou.myqcloud.com" {
		t.Fatalf("bucket url=%q, want %q", got, "https://shinerio-1258556983.cos.ap-guangzhou.myqcloud.com")
	}
}
