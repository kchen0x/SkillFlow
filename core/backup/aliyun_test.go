//go:build !provider_select || backup_aliyun

package backup

import "testing"

func TestNormalizeAliyunBucketName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain bucket", input: "my-skillflow-bucket", want: "my-skillflow-bucket"},
		{name: "trimmed bucket", input: "  my-skillflow-bucket  ", want: "my-skillflow-bucket"},
		{name: "full host", input: "my-skillflow-bucket.oss-cn-hangzhou.aliyuncs.com", want: "my-skillflow-bucket"},
		{name: "full url", input: "https://my-skillflow-bucket.oss-cn-hangzhou.aliyuncs.com/skillflow/", want: "my-skillflow-bucket"},
		{name: "endpoint only", input: "oss-cn-hangzhou.aliyuncs.com", wantErr: true},
		{name: "uppercase invalid", input: "MyBucket", wantErr: true},
		{name: "empty invalid", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAliyunBucketName(tt.input)
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
				t.Fatalf("normalizeAliyunBucketName(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeAliyunEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain endpoint", input: "oss-cn-hangzhou.aliyuncs.com", want: "oss-cn-hangzhou.aliyuncs.com"},
		{name: "full url", input: "https://oss-cn-hangzhou.aliyuncs.com", want: "oss-cn-hangzhou.aliyuncs.com"},
		{name: "bucket host", input: "https://my-skillflow-bucket.oss-cn-hangzhou.aliyuncs.com/", want: "oss-cn-hangzhou.aliyuncs.com"},
		{name: "empty invalid", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAliyunEndpoint(tt.input)
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
				t.Fatalf("normalizeAliyunEndpoint(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
