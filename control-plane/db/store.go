package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
)

type DBStore struct {
	DB *sql.DB
}

func NewDBStore(cfg *config.Config) (*DBStore, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// ping with timeout
	if err := pingWithTimeout(db, 3*time.Second); err != nil {
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}

	// pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	log.Println("connexion réussie à PostgreSQL")

	return &DBStore{DB: db}, nil
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- db.Ping()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("ping timeout after %v", timeout)
	}
}

func (s *DBStore) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}
