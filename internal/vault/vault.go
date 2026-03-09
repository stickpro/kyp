package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stickpro/kyp/internal/crypto"
	"github.com/stickpro/kyp/internal/models"
	"github.com/stickpro/kyp/internal/storage/sqlite"
	"github.com/stickpro/kyp/internal/storage/sqlite/repo_vault"
)

type Vault struct {
	meta      *models.VaultMetum
	masterKey []byte
	storage   sqlite.ILocalStorage
}

func Init(storage sqlite.ILocalStorage) *Vault {
	return &Vault{
		storage: storage,
	}
}

func (v *Vault) Create(ctx context.Context, password, name string) error {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	key := crypto.DeriveKey([]byte(password), salt)

	verifier, err := crypto.NewVerifier(key)
	if err != nil {
		return fmt.Errorf("create verifier: %w", err)
	}

	meta, err := v.storage.Vault().Create(ctx, repo_vault.CreateParams{
		ID:        uuid.New().String(),
		Name:      name,
		Salt:      salt,
		Verifier:  verifier,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	})

	if err != nil {
		return fmt.Errorf("create vault meta: %w", err)
	}

	v.meta = meta
	v.masterKey = key

	return nil
}

func (v *Vault) Open(ctx context.Context, password, name string) error {
	meta, err := v.storage.Vault().GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("get vault meta: %w", err)
	}
	key := crypto.DeriveKey([]byte(password), meta.Salt)
	if !crypto.CheckVerifier(meta.Verifier, key) {
		return fmt.Errorf("invalid master password")
	}
	v.masterKey = key
	v.meta = meta

	return nil
}

func (v *Vault) IsInitialized(ctx context.Context) bool {
	meta, err := v.storage.Vault().GetAll(ctx)
	return err == nil && len(meta) > 0
}

func (v *Vault) ListVaults(ctx context.Context) ([]*models.VaultMetum, error) {
	return v.storage.Vault().GetAll(ctx)
}

func (v *Vault) OpenByName(ctx context.Context, password, name string) error {
	return v.Open(ctx, password, name)
}

func (v *Vault) Close() {
	clear(v.masterKey)
	v.masterKey = nil
	v.meta = nil
}
