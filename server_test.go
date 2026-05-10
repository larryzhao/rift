package rye

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestParseShadowsocksServer(t *testing.T) {
	tests := []struct {
		name      string
		serverStr string
		want      *Server
	}{
		{
			name: "sip002 encoded userinfo",
			serverStr: "ss://" +
				base64.RawURLEncoding.EncodeToString([]byte("2022-blake3-aes-128-gcm:secret")) +
				"@example.com:8388#example",
			want: &Server{
				Protocol:   ProtoclSS,
				Name:       "example",
				Host:       "example.com",
				Port:       8388,
				User:       "secret",
				Encryption: "2022-blake3-aes-128-gcm",
			},
		},
		{
			name: "legacy encoded url",
			serverStr: "ss://" +
				base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:password@example.org:443")) +
				"#legacy",
			want: &Server{
				Protocol:   ProtoclSS,
				Name:       "legacy",
				Host:       "example.org",
				Port:       443,
				User:       "password",
				Encryption: "aes-256-gcm",
			},
		},
		{
			name:      "plain userinfo",
			serverStr: "ss://chacha20-ietf-poly1305:password@example.net:8443",
			want: &Server{
				Protocol:   ProtoclSS,
				Name:       "example.net",
				Host:       "example.net",
				Port:       8443,
				User:       "password",
				Encryption: "chacha20-ietf-poly1305",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServer(tt.serverStr)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if *got != *tt.want {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestParseVLessServerFromURL(t *testing.T) {
	got, err := ParseServer("vless://user@example.com:443?type=tcp&encryption=none&security=tls&sni=server.example#vless-example")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	want := &Server{
		Protocol:          ProtoclVLess,
		Name:              "vless-example",
		Host:              "example.com",
		Port:              443,
		User:              "user",
		Encryption:        "none",
		TransportProtocol: TransportProtocolTCP,
		ServerName:        "server.example",
		AllowInsecure:     true,
		Security:          "tls",
	}
	if *got != *want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestParseVMessServerFromBase64EncodedJSON(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte(`{
		"ps":"vmess-example",
		"add":"example.com",
		"port":"443",
		"id":"user-id",
		"aid":"0",
		"net":"tcp",
		"type":"none"
	}`))

	got, err := ParseServer(fmt.Sprintf("vmess://%s", encoded))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	want := &Server{
		Protocol:          ProtoclVMess,
		Name:              "vmess-example",
		Host:              "example.com",
		Port:              443,
		User:              "user-id",
		Encryption:        "auto",
		TransportProtocol: TransportProtocolTCP,
		Security:          "none",
		AlterID:           "0",
	}
	if *got != *want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestParseHysteria2ServerFromURL(t *testing.T) {
	got, err := ParseServer("hysteria2://user@example.com:443#hy-example")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	want := &Server{
		Protocol:      ProtoclHysteria2,
		Name:          "hy-example",
		Host:          "example.com",
		Port:          443,
		User:          "user",
		AllowInsecure: true,
	}
	if *got != *want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}
