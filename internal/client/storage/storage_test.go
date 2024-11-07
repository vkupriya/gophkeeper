package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vkupriya/gophkeeper/internal/client/models"
)

var dbpath = "/tmp/test.db"

func TestInitDB(t *testing.T) {
	store, err := NewSQLiteDB(dbpath)
	if err != nil {
		t.Errorf("failed to create local sqlite db: %v", err)
	}
	defer func() {
		if err := store.DB.Close(); err != nil {
			t.Error("failed to close local DB")
		}
	}()

	err = RunMigrations(store)
	require.NoError(t, err)
}

func TestSecretAdd(t *testing.T) {
	store, err := NewSQLiteDB(dbpath)
	if err != nil {
		t.Errorf("failed to open local sqlite db: %v", err)
	}
	defer func() {
		if err := store.DB.Close(); err != nil {
			t.Error("failed to close local DB")
		}
	}()

	secret := &models.Secret{
		Name:    "card01",
		Type:    "card",
		Meta:    "metadata",
		Data:    []byte("hello world"),
		Version: 1,
	}

	err = store.SecretAdd(secret)
	require.NoError(t, err)
}

func TestSecretGet(t *testing.T) {
	store, err := NewSQLiteDB(dbpath)
	if err != nil {
		t.Errorf("failed to open local sqlite db: %v", err)
	}
	defer func() {
		if err := store.DB.Close(); err != nil {
			t.Error("failed to close local DB")
		}
	}()

	secret := &models.Secret{
		Name:    "card01",
		Type:    "card",
		Meta:    "metadata",
		Data:    []byte("hello world"),
		Version: 1,
	}

	secretDB, err := store.SecretGet("card01")
	require.NoError(t, err)
	require.Equal(t, secretDB, secret)
}

func TestSecretList(t *testing.T) {
	store, err := NewSQLiteDB(dbpath)
	if err != nil {
		t.Errorf("failed to open local sqlite db: %v", err)
	}
	defer func() {
		if err := store.DB.Close(); err != nil {
			t.Error("failed to close local DB")
		}
	}()

	secretList, err := store.SecretList()
	require.NoError(t, err)
	if len(secretList) == 0 {
		t.Error("expected non-empty list of secrets")
	}
}

func TestSecretDeleteAll(t *testing.T) {
	store, err := NewSQLiteDB(dbpath)
	if err != nil {
		t.Errorf("failed to open local sqlite db: %v", err)
	}
	defer func() {
		if err := store.DB.Close(); err != nil {
			t.Error("failed to close local DB")
		}
	}()

	err = store.SecretDeleteAll()
	require.NoError(t, err)
}
