//nolint:all
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stickpro/kyp/internal/storage/sqlite"
	"github.com/stickpro/kyp/internal/vault"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: kyp-import <csv-file> <db-path> <master-password> [folder]\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  kyp-import bitwarden.csv kyp.db mypassword\n")
		fmt.Fprintf(os.Stderr, "  kyp-import bitwarden.csv kyp.db mypassword \"My Folder\"\n")
		os.Exit(1)
	}

	csvPath := os.Args[1]
	dbPath := os.Args[2]
	password := os.Args[3]

	var folderFilter string
	if len(os.Args) >= 5 {
		folderFilter = os.Args[4]
	}

	storage, err := sqlite.InitLocalStorage(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer storage.Close()

	v := vault.Init(storage)
	if err := v.Open(context.Background(), password, "default"); err != nil {
		return fmt.Errorf("open vault: %w", err)
	}
	defer v.Close()

	f, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	records, err := parseCSV(f)
	if err != nil {
		return fmt.Errorf("parse csv: %w", err)
	}

	imported, skipped := 0, 0
	for _, r := range records {
		if r.typ != "" && r.typ != "login" {
			skipped++
			continue
		}

		if folderFilter != "" && !strings.EqualFold(r.folder, folderFilter) {
			skipped++
			continue
		}

		dto := vault.EntryDTO{
			Title:    r.name,
			Username: strPtr(r.username),
			Password: strPtr(r.password),
			URL:      strPtr(r.uri),
			Notes:    strPtr(r.notes),
			// TOTP defaults
			TOTPAlgorithm: "SHA1",
			TOTPDigits:    6,
			TOTPPeriod:    30,
		}
		if r.totp != "" {
			secret := parseTOTPSecret(r.totp)
			dto.TOTPSecret = &secret
		}

		if err := v.CreateEntry(context.Background(), dto); err != nil {
			fmt.Fprintf(os.Stderr, "skip %q: %v\n", r.name, err)
			skipped++
			continue
		}
		fmt.Printf("  + %s\n", r.name)
		imported++
	}

	fmt.Printf("\nDone: %d imported, %d skipped\n", imported, skipped)
	return nil
}

// Bitwarden CSV columns:
// folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp
type bitwardenRow struct {
	folder   string
	typ      string
	name     string
	notes    string
	uri      string
	username string
	password string
	totp     string
}

func parseCSV(r io.Reader) ([]bitwardenRow, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	idx := buildIndex(header)

	var rows []bitwardenRow
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		rows = append(rows, bitwardenRow{
			folder:   get(rec, idx, "folder"),
			typ:      get(rec, idx, "type"),
			name:     get(rec, idx, "name"),
			notes:    get(rec, idx, "notes"),
			uri:      get(rec, idx, "login_uri"),
			username: get(rec, idx, "login_username"),
			password: get(rec, idx, "login_password"),
			totp:     get(rec, idx, "login_totp"),
		})
	}

	return rows, nil
}

func buildIndex(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.TrimSpace(strings.ToLower(h))] = i
	}
	return m
}

func get(rec []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

// parseTOTPSecret извлекает секрет из otpauth URI или возвращает строку как есть.
// otpauth://totp/label?secret=XXXX&issuer=...
func parseTOTPSecret(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "otpauth://") {
		return raw
	}
	if idx := strings.Index(raw, "secret="); idx != -1 {
		s := raw[idx+7:]
		if end := strings.IndexAny(s, "&"); end != -1 {
			return s[:end]
		}
		return s
	}
	return raw
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
