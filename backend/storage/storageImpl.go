package storage

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

type storage struct {  
    db *sql.DB
}

func NewStorage(cfg *Config) (Storage, error) {
    db, err := sql.Open("postgres", cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    return &storage{db: db}, nil  
}

func (s *storage) Close() error {
    return s.db.Close()
}

func (s *storage) Ping() error {
    return s.db.Ping()
}