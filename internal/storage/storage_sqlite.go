package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type SQLiteStorage struct {
	db *sql.DB

	salt string
}

func NewSQLiteStorage(storagePath string) (*SQLiteStorage, error) {
	_, err := os.Stat(filepath.Join(storagePath, "storage.db"))

	var db *sql.DB
	if os.IsNotExist(err) {
		scheme, err := os.ReadFile(filepath.Join(storagePath, "scheme.sql"))
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot init SQLite storage: %w", err)
		}

		db, err = sql.Open("sqlite3", filepath.Join(storagePath, "storage.db"))
		if err != nil {
			return nil, fmt.Errorf("cannot open sqlite3 connection: %w", err)
		}

		_, err = db.Exec(string(scheme))
		if err != nil {
			db.Close()
			os.Remove(filepath.Join(storagePath, "storage.db"))

			return nil, fmt.Errorf("cannot create DB schema: %w", err)
		}
	} else {
		db, err = sql.Open("sqlite3", filepath.Join(storagePath, "storage.db"))
		if err != nil {
			return nil, fmt.Errorf("cannot open sqlite3 connection: %w", err)
		}
	}

	return &SQLiteStorage{
		db:   db,
		salt: "storageSalt14", // TODO: move salt to config
	}, nil
}

func (s *SQLiteStorage) GetClient(ctx context.Context, login, password string) (*Client, error) {
	hashPassword := s.hashPassword(password)
	var l, p string
	err := s.db.QueryRowContext(ctx, "SELECT login, password FROM clients WHERE login = $1 AND password = $2", login, hashPassword).Scan(&l, &p)
	if err != nil && errors.Is(sql.ErrNoRows, err) {
		return nil, ErrClientNotFound
	} else if err != nil {
		return nil, fmt.Errorf("cannot get client data: %w", err)
	}

	return &Client{
		Login:    l,
		password: p,
	}, nil
}

func (s *SQLiteStorage) CreateClient(ctx context.Context, login, password string) (*Client, error) {
	hashPassword := s.hashPassword(password)

	_, err := s.db.ExecContext(ctx, "INSERT INTO clients (login, password) VALUES ($1, $2)", login, hashPassword)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}

	return &Client{
		Login:    login,
		password: hashPassword,
	}, nil
}

func (s *SQLiteStorage) hashPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password+s.salt)))
}
