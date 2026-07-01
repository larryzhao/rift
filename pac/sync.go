package pac

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GFWListURL is the upstream base64-encoded gfwlist, refreshed ~every 6h.
const GFWListURL = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

// GFWListSyncResult reports the outcome of a SyncGFWList call.
type GFWListSyncResult struct {
	// Changed is true only when outFile was rewritten with new content.
	Changed bool
	// ETag is the server-provided ETag to replay as If-None-Match next time.
	ETag string
	// SHA256 is the hex content hash of the current gfwlist.
	SHA256 string
	// Bytes is the size of the downloaded body (0 when not modified).
	Bytes int
}

// SyncGFWList downloads the upstream gfwlist (still base64) and writes it to
// outFile only when the content actually changed. prevETag is sent as
// If-None-Match so an unchanged upstream returns 304 without a body; prevSHA256
// guards against rewriting the file when a 200 response carries identical
// content (e.g. only the embedded timestamp changed).
func SyncGFWList(outFile, prevETag, prevSHA256 string) (GFWListSyncResult, error) {
	req, err := http.NewRequest(http.MethodGet, GFWListURL, nil)
	if err != nil {
		return GFWListSyncResult{}, fmt.Errorf("build gfwlist request: %w", err)
	}
	if prevETag != "" {
		req.Header.Set("If-None-Match", prevETag)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GFWListSyncResult{}, fmt.Errorf("download gfwlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return GFWListSyncResult{Changed: false, ETag: prevETag, SHA256: prevSHA256}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return GFWListSyncResult{}, fmt.Errorf("download gfwlist: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GFWListSyncResult{}, fmt.Errorf("read gfwlist body: %w", err)
	}

	sum := sha256.Sum256(body)
	res := GFWListSyncResult{
		ETag:   resp.Header.Get("ETag"),
		SHA256: hex.EncodeToString(sum[:]),
		Bytes:  len(body),
	}

	// Content identical to what we already have: keep the ETag fresh but skip
	// the rewrite.
	if res.SHA256 == prevSHA256 {
		return res, nil
	}

	if err := os.WriteFile(outFile, body, 0o644); err != nil {
		return GFWListSyncResult{}, fmt.Errorf("write gfwlist file: %w", err)
	}
	res.Changed = true
	return res, nil
}
