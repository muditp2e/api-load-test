// vault_cert_expiry.go
//
// Reads keys from keys.txt (one per line). Supported line formats:
//
// 1) Plain key (will call Vault API):
//    kwl-abc123...
//
// 2) Inline cert (NO API call):
//    kwl-abc123="-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----"
//
// Notes:
// - Lines starting with # are ignored (comments).
// - Empty lines are ignored.
// - Inline cert supports escaped sequences like \n via strconv.Unquote.
// - Output CSV columns: id, environment, key, expiry, keyFetchedFromVault
// - Output filename includes env: vault_cert_expiry_<env>.csv
//
// Usage:
//   go run vault_cert_expiry.go -env=staging
//   go run vault_cert_expiry.go -env=staging -keys=keys.txt -url=http://localhost:9000/vault/getVaultKey

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
	"strconv"
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

type entry struct {
	key         string
	inlineCert  string // if non-empty, use this cert and skip API
	fromVault   bool   // true if fetched from Vault API; false if inlineCert provided
	originalRow string // for debug
}

func main() {
	env := flag.String("env", "mainnet", "Environment name (written to CSV and used in output filename)")
	keysPath := flag.String("keys", "mainnet.txt", "Path to keys file (one key per line)")
	url := flag.String("url", "http://localhost:9000/vault/getVaultKey", "Vault API URL")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP timeout per request")
	flag.Parse()

	if strings.TrimSpace(*env) == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -env is required (e.g., -env=staging)")
		os.Exit(1)
	}

	entries, err := readEntries(*keysPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading keys: %v\n", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: no usable lines found in keys file")
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
	if err := w.Write([]string{"id", "environment", "key", "expiry", "keyFetchedFromVault"}); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: writing CSV header: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: *timeout}

	for i, e := range entries {
		expiryStr := ""
		fetchedFromVault := e.fromVault

		if e.inlineCert != "" {
			// Parse inline cert (NO API call)
			notAfter, err := parseCertExpiry(e.inlineCert)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARN: inline cert parse failed for key=%q: %v\n", e.key, err)
			} else {
				expiryStr = notAfter.UTC().Format(time.RFC3339)
			}
		} else {
			// Fetch cert from Vault API
			certPEM, err := fetchCertificate(context.Background(), client, *url, e.key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARN: vault fetch failed for key=%q: %v\n", e.key, err)
			} else {
				notAfter, err := parseCertExpiry(certPEM)
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: fetched cert parse failed for key=%q: %v\n", e.key, err)
				} else {
					expiryStr = notAfter.UTC().Format(time.RFC3339)
				}
			}
		}

		row := []string{
			fmt.Sprintf("%d", i+1),
			*env,
			e.key,
			expiryStr,
			strconv.FormatBool(fetchedFromVault),
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

	fmt.Printf("Wrote %d rows to %s\n", len(entries), outPath)
}

func readEntries(path string) ([]entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []entry
	sc := bufio.NewScanner(f)

	// Increase scanner buffer for long inline cert lines.
	// (Default 64K can be too small for large PEM chains.)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 10*1024*1024) // up to 10MB line

	for sc.Scan() {
		raw := sc.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		e, ok, err := parseLine(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: skipping malformed line: %q err=%v\n", line, err)
			continue
		}
		if !ok {
			continue
		}
		e.originalRow = line
		entries = append(entries, e)
	}
	return entries, sc.Err()
}

// parseLine supports:
// - key
// - key="...cert..."
// - "key"="...cert..." (optional quoting)
func parseLine(line string) (entry, bool, error) {
	// Inline cert format contains '='
	if strings.Contains(line, "=") {
		parts := strings.SplitN(line, "=", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		key, err := stripOptionalQuotes(left)
		if err != nil {
			return entry{}, false, fmt.Errorf("invalid key quoting: %w", err)
		}
		if key == "" {
			return entry{}, false, fmt.Errorf("empty key")
		}

		// Right side must be a quoted string if it contains spaces/newlines escapes etc.
		// We accept either:
		//  - "...."
		//  - '....' (rare)
		// If not quoted, treat it as a literal cert text.
		cert, err := parseMaybeQuotedString(right)
		if err != nil {
			return entry{}, false, fmt.Errorf("invalid cert string: %w", err)
		}
		cert = normalizePEMString(cert)

		return entry{
			key:        key,
			inlineCert: cert,
			fromVault:  false, // explicitly not fetched from vault
		}, true, nil
	}

	// Plain key format
	key := strings.TrimSpace(line)
	key, err := stripOptionalQuotes(key)
	if err != nil {
		return entry{}, false, fmt.Errorf("invalid key quoting: %w", err)
	}
	if key == "" {
		return entry{}, false, nil
	}
	return entry{
		key:       key,
		fromVault: true,
	}, true, nil
}

func stripOptionalQuotes(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			// Use Unquote to also process escapes safely
			return strconv.Unquote(s)
		}
	}
	return s, nil
}

func parseMaybeQuotedString(s string) (string, error) {
	s = strings.TrimSpace(s)
	// If it's quoted, unquote (handles \n, \t, \", etc.)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return strconv.Unquote(s)
	}
	// Otherwise use as-is
	return s, nil
}

func normalizePEMString(cert string) string {
	// If someone stored literal \n without quoting, convert common patterns.
	// (If it was quoted, Unquote already converts \n to newlines.)
	cert = strings.ReplaceAll(cert, `\r\n`, "\n")
	cert = strings.ReplaceAll(cert, `\n`, "\n")
	cert = strings.ReplaceAll(cert, "\r", "\n")
	return strings.TrimSpace(cert)
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
		return "", fmt.Errorf("non-2xx status: %s body=%s", resp.Status, truncate(string(b), 400))
	}

	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err != nil {
		return "", fmt.Errorf("unmarshal response: %w body=%s", err, truncate(string(b), 400))
	}

	cert := strings.TrimSpace(ar.VaultResponse.Credentials.Certificate)
	if cert == "" {
		return "", fmt.Errorf("certificate field empty/missing")
	}
	// json.Unmarshal already unescapes \n into actual newlines.
	return strings.TrimSpace(cert), nil
}

func parseCertExpiry(certPEM string) (time.Time, error) {
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
		if len(rest) == 0 {
			return time.Time{}, fmt.Errorf("no CERTIFICATE PEM block found")
		}
	}
}

func sanitizeFilename(s string) string {
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
