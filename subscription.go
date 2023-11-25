package rye

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
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

	encodedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(encodedData))
	if err != nil {
		return nil, err
	}

	var servers []*Server
	scanner := bufio.NewScanner(bytes.NewReader(decodedData))
	for scanner.Scan() {
		url := scanner.Text()
		server, err := ParseServerFromURL(url)
		if err != nil {
			PrintVerbose("parse server url %s err: %s", url, err.Error())
			continue
		}
		servers = append(servers, server)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}
