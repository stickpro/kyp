package gui

import (
	"context"
	"time"

	"github.com/stickpro/kyp/internal/crypto"
	"github.com/stickpro/kyp/internal/storage/sqlite"
	"github.com/stickpro/kyp/internal/totp"
	"github.com/stickpro/kyp/internal/vault"
)

// EntryDTO is the entry representation exposed to the Svelte frontend.
type EntryDTO struct {
	ID            *string `json:"id"`
	Title         string  `json:"title"`
	Username      *string `json:"username"`
	Password      *string `json:"password"`
	URL           *string `json:"url"`
	Notes         *string `json:"notes"`
	TOTPSecret    *string `json:"totpSecret"`
	TOTPIssuer    *string `json:"totpIssuer"`
	TOTPDigits    int64   `json:"totpDigits"`
	TOTPPeriod    int64   `json:"totpPeriod"`
	TOTPAlgorithm string  `json:"totpAlgorithm"`
}

// TOTPState is returned by GetTOTPCode — current code + seconds remaining.
type TOTPState struct {
	Code      string `json:"code"`
	Remaining int    `json:"remaining"`
}

// App holds application state and exposes methods to the Wails frontend.
type App struct {
	ctx     context.Context
	vault   *vault.Vault
	storage sqlite.ILocalStorage
}

func NewApp(storage sqlite.ILocalStorage) *App {
	return &App{
		vault:   vault.Init(storage),
		storage: storage,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {
	a.vault.Close()
}

func (a *App) HasVault() bool {
	vaults, err := a.vault.ListVaults(a.ctx)
	return err == nil && len(vaults) > 0
}

func (a *App) ListVaults() ([]string, error) {
	vaults, err := a.vault.ListVaults(a.ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(vaults))
	for i, v := range vaults {
		names[i] = v.Name
	}
	return names, nil
}

func (a *App) CreateVault(name, password string) error {
	return a.vault.Create(a.ctx, password, name)
}

func (a *App) Unlock(name, password string) error {
	return a.vault.OpenByName(a.ctx, password, name)
}

func (a *App) Lock() {
	a.vault.Close()
}

func (a *App) ListEntries() ([]EntryDTO, error) {
	entries, err := a.vault.ListEntries(a.ctx, 1000, 0)
	if err != nil {
		return nil, err
	}
	result := make([]EntryDTO, len(entries))
	for i, e := range entries {
		result[i] = toDTO(e)
	}
	return result, nil
}

func (a *App) GetEntry(id string) (*EntryDTO, error) {
	e, err := a.vault.GetEntry(a.ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toDTO(*e)
	return &dto, nil
}

func (a *App) CreateEntry(dto EntryDTO) error {
	return a.vault.CreateEntry(a.ctx, fromDTO(dto))
}

func (a *App) UpdateEntry(id string, dto EntryDTO) error {
	return a.vault.UpdateEntry(a.ctx, id, fromDTO(dto))
}

func (a *App) DeleteEntry(id string) error {
	return a.vault.DeleteEntry(a.ctx, id)
}

func (a *App) GeneratePassword(length int) (string, error) {
	return crypto.GeneratePassword(length, true, true, true)
}

// GetTOTPCode generates the current TOTP code for an entry.
func (a *App) GetTOTPCode(id string) (*TOTPState, error) {
	e, err := a.vault.GetEntry(a.ctx, id)
	if err != nil {
		return nil, err
	}
	if e.TOTPSecret == nil || *e.TOTPSecret == "" {
		return nil, nil
	}
	period := int(e.TOTPPeriod)
	if period == 0 {
		period = 30
	}
	digits := int(e.TOTPDigits)
	if digits == 0 {
		digits = 6
	}
	now := time.Now()
	code, err := totp.Code(*e.TOTPSecret, now, digits, period)
	if err != nil {
		return nil, err
	}
	return &TOTPState{
		Code:      code,
		Remaining: totp.Remaining(period, now),
	}, nil
}

func toDTO(e vault.EntryDTO) EntryDTO {
	return EntryDTO{
		ID:            e.ID,
		Title:         e.Title,
		Username:      e.Username,
		Password:      e.Password,
		URL:           e.URL,
		Notes:         e.Notes,
		TOTPSecret:    e.TOTPSecret,
		TOTPIssuer:    e.TOTPIssuer,
		TOTPDigits:    e.TOTPDigits,
		TOTPPeriod:    e.TOTPPeriod,
		TOTPAlgorithm: e.TOTPAlgorithm,
	}
}

func fromDTO(dto EntryDTO) vault.EntryDTO {
	digits := dto.TOTPDigits
	if digits == 0 {
		digits = 6
	}
	period := dto.TOTPPeriod
	if period == 0 {
		period = 30
	}
	algo := dto.TOTPAlgorithm
	if algo == "" {
		algo = "SHA1"
	}
	return vault.EntryDTO{
		Title:         dto.Title,
		Username:      dto.Username,
		Password:      dto.Password,
		URL:           dto.URL,
		Notes:         dto.Notes,
		TOTPSecret:    dto.TOTPSecret,
		TOTPIssuer:    dto.TOTPIssuer,
		TOTPDigits:    digits,
		TOTPPeriod:    period,
		TOTPAlgorithm: algo,
	}
}
