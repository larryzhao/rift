package rift

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Subscription struct {
	Name          string    `yaml:"name"`
	URL           string    `yaml:"url"`
	AddedAt       time.Time `yaml:"added_at"`
	LastUpdatedAt time.Time `yaml:"last_updated_at"`
	SkipUpdate    bool      `yaml:"skip_update"`
}

func (sub *Subscription) Fetch(ctx context.Context) ([]*Server, error) {
	var err error

	resp, err := http.Get(sub.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	encodedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decodedData, err := decodeSubscriptionData(encodedData)
	if err != nil {
		return nil, err
	}

	var servers []*Server
	scanner := bufio.NewScanner(bytes.NewReader(decodedData))
	for scanner.Scan() {
		url := scanner.Text()
		server, err := ParseServer(url)
		if err != nil {
			PrintlnVerbose("parse server url %s err: %s", url, err.Error())
			continue
		}
		servers = append(servers, server)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func decodeSubscriptionData(encodedData []byte) ([]byte, error) {
	encoded := string(bytes.TrimSpace(encodedData))

	decodedData, stdErr := base64.StdEncoding.DecodeString(encoded)
	if stdErr == nil {
		return decodedData, nil
	}

	decodedData, rawURLErr := base64.RawURLEncoding.DecodeString(encoded)
	if rawURLErr == nil {
		return decodedData, nil
	}

	PrintlnError("decode subscription data err: std encoding: %s; raw url encoding: %s", stdErr.Error(), rawURLErr.Error())
	return nil, fmt.Errorf("decode subscription data err: std encoding: %w; raw url encoding: %w", stdErr, rawURLErr)
}
