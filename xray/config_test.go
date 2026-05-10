package xray

import (
	"encoding/json"
	"testing"

	"github.com/larryzhao/rye"
)

func TestToShadowsocksOutbound(t *testing.T) {
	outbound, err := toOutbound(&rye.Server{
		Protocol:   rye.ProtoclSS,
		Host:       "example.com",
		Port:       8388,
		User:       "secret",
		Encryption: "aes-256-gcm",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if outbound.Protocol != "shadowsocks" {
		t.Fatalf("expected protocol shadowsocks, got %s", outbound.Protocol)
	}

	var settings struct {
		Address  string `json:"address"`
		Port     int    `json:"port"`
		Method   string `json:"method"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(*outbound.Settings, &settings); err != nil {
		t.Fatalf("unmarshal settings err: %v", err)
	}

	if settings.Address != "example.com" || settings.Port != 8388 || settings.Method != "aes-256-gcm" || settings.Password != "secret" {
		t.Fatalf("unexpected shadowsocks settings: %+v", settings)
	}
}
