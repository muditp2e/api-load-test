// vault_cert_expiry.go
//
// Usage:
//   go run vault_cert_expiry.go -env=staging
//   go run vault_cert_expiry.go -env=staging -keys=keys.txt -url=http://localhost:9000/vault/getVaultKey
//
// What it does:
// - Reads keys from keys.txt (one per line; empty lines/# comments ignored)
// - Calls POST <url> with JSON {"vaultKey":"<key>"}
// - Extracts: vaultResponse.credentials.certificate (PEM string)
// - Parses X.509 cert and gets NotAfter (expiry)
// - Writes CSV: vault_cert_expiry_<env>.csv with columns: id, environment, key, expiry

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"encoding/csv"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type apiRequest struct {
	VaultKey string `json:"vaultKey"`
}

type apiResponse struct {
	VaultResponse struct {
		Credentials struct {
			Certificate string `json:"certificate"`
		} `json:"credentials"`
	} `json:"vaultResponse"`
}

func main() {
	env := flag.String("env", "mainnet", "Environment name (will be written to CSV and used in output filename)")
	keysPath := flag.String("keys", "mainnet.txt", "Path to keys file (one key per line)")
	url := flag.String("url", "http://localhost:9000/vault/getVaultKey", "Vault API URL")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP timeout per request")
	flag.Parse()

	if strings.TrimSpace(*env) == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -env is required (e.g., -env=staging)")
		os.Exit(1)
	}

	keys, err := readKeys(*keysPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading keys: %v\n", err)
		os.Exit(1)
	}
	if len(keys) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: no keys found in keys file")
		os.Exit(1)
	}

	outName := fmt.Sprintf("vault_cert_expiry_%s.csv", sanitizeFilename(*env))
	outPath := filepath.Join(".", outName)

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: creating CSV: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"id", "environment", "key", "expiry"}); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: writing CSV header: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: *timeout}

	for i, k := range keys {
		expiryStr := ""

		certPEM, err := fetchCertificate(context.Background(), client, *url, k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: key=%q fetch failed: %v\n", k, err)
		} else {
			notAfter, err := parseCertExpiry(certPEM)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARN: key=%q cert parse failed: %v\n", k, err)
			} else {
				// Use RFC3339 for unambiguous timestamps
				expiryStr = notAfter.UTC().Format(time.RFC3339)
			}
		}

		row := []string{
			fmt.Sprintf("%d", i+1),
			*env,
			k,
			expiryStr,
		}
		if err := w.Write(row); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing CSV row (id=%d): %v\n", i+1, err)
			os.Exit(1)
		}
	}

	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: flushing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %d rows to %s\n", len(keys), outPath)
}

func readKeys(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var keys []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keys = append(keys, line)
	}
	return keys, sc.Err()
}

func fetchCertificate(ctx context.Context, client *http.Client, url, key string) (string, error) {
	reqBody, err := json.Marshal(apiRequest{VaultKey: key})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// include body for debugging
		return "", fmt.Errorf("non-2xx status: %s body=%s", resp.Status, truncate(string(b), 400))
	}

	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err != nil {
		return "", fmt.Errorf("unmarshal response: %w body=%s", err, truncate(string(b), 400))
	}

	cert := ar.VaultResponse.Credentials.Certificate
	if strings.TrimSpace(cert) == "" {
		return "", fmt.Errorf("certificate field empty/missing")
	}

	// cert is already unescaped by json.Unmarshal (e.g., \n becomes actual newlines).
	return cert, nil
}

func parseCertExpiry(certPEM string) (time.Time, error) {
	// Expect PEM. If multiple PEM blocks exist, pick the first CERTIFICATE block.
	rest := []byte(certPEM)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return time.Time{}, fmt.Errorf("no PEM block found")
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return time.Time{}, fmt.Errorf("x509 parse: %w", err)
			}
			return cert.NotAfter, nil
		}
		// Continue if it's some other PEM block type.
		if len(rest) == 0 {
			return time.Time{}, fmt.Errorf("no CERTIFICATE PEM block found")
		}
	}
}

func sanitizeFilename(s string) string {
	// Keep it simple: replace path separators and spaces.
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, " ", "_")
	if s == "" {
		return "env"
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
