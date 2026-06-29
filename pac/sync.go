package pac

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GFWListURL is the upstream base64-encoded gfwlist, refreshed ~every 6h.
const GFWListURL = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

// SyncGFWList downloads the upstream gfwlist and writes it (still base64) to
// outFile. GeneratePAC decodes it at serve time.
func SyncGFWList(outFile string) (int, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(GFWListURL)
	if err != nil {
		return 0, fmt.Errorf("download gfwlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download gfwlist: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read gfwlist body: %w", err)
	}

	if err := os.WriteFile(outFile, body, 0o644); err != nil {
		return 0, fmt.Errorf("write gfwlist file: %w", err)
	}
	return len(body), nil
}
