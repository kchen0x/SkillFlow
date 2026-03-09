//go:build !provider_select || backup_huawei

package backup

import "testing"

func TestNormalizeHuaweiBucketName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain bucket", input: "my-skillflow-bucket", want: "my-skillflow-bucket"},
		{name: "trimmed bucket", input: "  my-skillflow-bucket  ", want: "my-skillflow-bucket"},
		{name: "full host", input: "my-skillflow-bucket.obs.cn-north-4.myhuaweicloud.com", want: "my-skillflow-bucket"},
		{name: "full url", input: "https://my-skillflow-bucket.obs.cn-north-4.myhuaweicloud.com/skillflow/", want: "my-skillflow-bucket"},
		{name: "endpoint only", input: "obs.cn-north-4.myhuaweicloud.com", wantErr: true},
		{name: "uppercase invalid", input: "MyBucket", wantErr: true},
		{name: "empty invalid", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeHuaweiBucketName(tt.input)
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
				t.Fatalf("normalizeHuaweiBucketName(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeHuaweiEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain endpoint", input: "obs.cn-north-4.myhuaweicloud.com", want: "obs.cn-north-4.myhuaweicloud.com"},
		{name: "full url", input: "https://obs.cn-north-4.myhuaweicloud.com", want: "obs.cn-north-4.myhuaweicloud.com"},
		{name: "bucket host", input: "https://my-skillflow-bucket.obs.cn-north-4.myhuaweicloud.com/", want: "obs.cn-north-4.myhuaweicloud.com"},
		{name: "empty invalid", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeHuaweiEndpoint(tt.input)
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
				t.Fatalf("normalizeHuaweiEndpoint(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
