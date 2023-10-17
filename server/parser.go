package server

import (
	"fmt"
	"net/url"
)

// ParseURL parse server from URL
// trojan://31b98cae-da2d-4456-b351-f91838313f0a@23.227.38.158:443?allowInsecure=0&peer=kus1.yfjc.sbs&sni=kus1.yfjc.sbs&type=ws&host=kus1.yfjc.sbs&path=/yfjc/kus1#%F0%9F%87%BA%F0%9F%87%B8%E7%BE%8E%E5%9B%BD1%E5%8F%B7-0.1%E5%80%8D
func ParseURL(u string) (Server, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("parse url %s err: %w", u, err)
	}

	switch parsedURL.Scheme {
	case "trojan":
		return parseTrojanFromURL(parsedURL)
	default:
		return nil, fmt.Errorf("scheme not supported: %s", parsedURL.Scheme)
	}
}
