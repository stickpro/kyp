package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stickpro/kyp/internal/crypto"
	"github.com/stickpro/kyp/internal/models"
	"github.com/stickpro/kyp/internal/storage/sqlite/repo_entry"
)

type EntryDTO struct {
	ID    *string
	Title string

	Username *string
	Password *string
	URL      *string
	Notes    *string

	// TOTP - secret and issuer nullable
	TOTPSecret    *string
	TOTPIssuer    *string
	TOTPAlgorithm string
	TOTPDigits    int64
	TOTPPeriod    int64
}

func (v *Vault) CreateEntry(ctx context.Context, d EntryDTO) error {
	if v.meta == nil {
		return NotOpen
	}

	enc, err := v.encryptEntry(d)
	if err != nil {
		return err
	}

	// set defaults for TOTP
	if d.TOTPAlgorithm == "" {
		d.TOTPAlgorithm = "SHA1"
	}
	if d.TOTPDigits == 0 {
		d.TOTPDigits = 6
	}
	if d.TOTPPeriod == 0 {
		d.TOTPPeriod = 30
	}

	params := repo_entry.CreateParams{
		ID:            uuid.New().String(),
		Title:         d.Title,
		Username:      enc.username,
		Password:      enc.password,
		Url:           enc.url,
		Notes:         enc.notes,
		TotpSecret:    enc.totpSecret,
		TotpIssuer:    d.TOTPIssuer,
		TotpAlgorithm: d.TOTPAlgorithm,
		TotpDigits:    d.TOTPDigits,
		TotpPeriod:    d.TOTPPeriod,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
	}

	err = v.storage.Entries().Create(ctx, params)
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}
	return nil
}

func (v *Vault) UpdateEntry(ctx context.Context, id string, d EntryDTO) error {
	if v.meta == nil {
		return NotOpen
	}

	enc, err := v.encryptEntry(d)
	if err != nil {
		return err
	}

	// set defaults for TOTP
	if d.TOTPAlgorithm == "" {
		d.TOTPAlgorithm = "SHA1"
	}
	if d.TOTPDigits == 0 {
		d.TOTPDigits = 6
	}
	if d.TOTPPeriod == 0 {
		d.TOTPPeriod = 30
	}

	params := repo_entry.UpdateParams{
		ID:            id,
		Title:         d.Title,
		Username:      enc.username,
		Password:      enc.password,
		Url:           enc.url,
		Notes:         enc.notes,
		TotpSecret:    enc.totpSecret,
		TotpIssuer:    d.TOTPIssuer,
		TotpAlgorithm: d.TOTPAlgorithm,
		TotpDigits:    d.TOTPDigits,
		TotpPeriod:    d.TOTPPeriod,
	}

	_, err = v.storage.Entries().Update(ctx, params)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	return nil
}

func (v *Vault) GetEntry(ctx context.Context, id string) (*EntryDTO, error) {
	if v.meta == nil {
		return nil, NotOpen
	}

	entry, err := v.storage.Entries().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	dto, err := v.decryptEntry(entry)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (v *Vault) ListEntries(ctx context.Context, limit, offset int64) ([]EntryDTO, error) {
	if v.meta == nil {
		return nil, NotOpen
	}
	entrys, err := v.storage.Entries().GetWithPaginate(ctx, repo_entry.GetWithPaginateParams{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, err
	}

	dtos := make([]EntryDTO, len(entrys))
	for i, entry := range entrys {
		dtos[i], err = v.decryptEntry(entry)
		if err != nil {
			return nil, err
		}
	}
	return dtos, nil
}

func (v *Vault) DeleteEntry(ctx context.Context, id string) error {
	if v.meta == nil {
		return NotOpen
	}
	return v.storage.Entries().Delete(ctx, id)
}

type encryptedFields struct {
	username   []byte
	password   []byte
	url        []byte
	notes      []byte
	totpSecret []byte
}

func (v *Vault) encryptEntry(d EntryDTO) (encryptedFields, error) {
	cUsername, err := encryptOptional(d.Username, v.masterKey)
	if err != nil {
		return encryptedFields{}, fmt.Errorf("encrypt username: %w", err)
	}
	cPassword, err := encryptOptional(d.Password, v.masterKey)
	if err != nil {
		return encryptedFields{}, fmt.Errorf("encrypt password: %w", err)
	}
	cURL, err := encryptOptional(d.URL, v.masterKey)
	if err != nil {
		return encryptedFields{}, fmt.Errorf("encrypt url: %w", err)
	}
	cNotes, err := encryptOptional(d.Notes, v.masterKey)
	if err != nil {
		return encryptedFields{}, fmt.Errorf("encrypt notes: %w", err)
	}
	cTOTPSecret, err := encryptOptional(d.TOTPSecret, v.masterKey)
	if err != nil {
		return encryptedFields{}, fmt.Errorf("encrypt totp secret: %w", err)
	}
	return encryptedFields{
		username:   cUsername,
		password:   cPassword,
		url:        cURL,
		notes:      cNotes,
		totpSecret: cTOTPSecret,
	}, nil
}

func encryptOptional(s *string, key []byte) ([]byte, error) {
	if s == nil {
		return nil, nil
	}
	return crypto.Encrypt([]byte(*s), key)
}

func decryptOptional(data []byte, key []byte) (*string, error) {
	if data == nil {
		return nil, nil
	}
	plain, err := crypto.Decrypt(data, key)
	if err != nil {
		return nil, err
	}
	s := string(plain)
	return &s, nil
}

func (v *Vault) decryptEntry(entry *models.Entry) (EntryDTO, error) {
	dUsername, err := decryptOptional(entry.Username, v.masterKey)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("decrypt username: %w", err)
	}
	dPassword, err := decryptOptional(entry.Password, v.masterKey)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("decrypt password: %w", err)
	}
	dURL, err := decryptOptional(entry.Url, v.masterKey)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("decrypt url: %w", err)
	}
	dNotes, err := decryptOptional(entry.Notes, v.masterKey)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("decrypt notes: %w", err)
	}
	dTOTPSecret, err := decryptOptional(entry.TotpSecret, v.masterKey)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("decrypt totp secret: %w", err)
	}

	return EntryDTO{
		ID:            &entry.ID,
		Title:         entry.Title,
		Username:      dUsername,
		Password:      dPassword,
		URL:           dURL,
		Notes:         dNotes,
		TOTPSecret:    dTOTPSecret,
		TOTPIssuer:    entry.TotpIssuer,
		TOTPAlgorithm: entry.TotpAlgorithm,
		TOTPDigits:    entry.TotpDigits,
		TOTPPeriod:    entry.TotpPeriod,
	}, nil
}
