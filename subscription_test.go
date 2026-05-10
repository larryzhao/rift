package rye

import (
	"encoding/base64"
	"testing"
)

func TestDecodeSubscriptionData(t *testing.T) {
	data := []byte("hysteria2://user@example.com:443#example\n")

	tests := []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{
			name:    "std encoding",
			encoded: base64.StdEncoding.EncodeToString(data),
		},
		{
			name:    "raw url encoding",
			encoded: base64.RawURLEncoding.EncodeToString(data),
		},
		{
			name:    "invalid encoding",
			encoded: "not base64 encoded",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeSubscriptionData([]byte(tt.encoded))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if string(got) != string(data) {
				t.Fatalf("expected %q, got %q", data, got)
			}
		})
	}
}
