package storage

import (
	"context"
	"database/sql"
)

type LogRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(ctx context.Context, entry LogEntry) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO logs(id, source, level, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		entry.ID, entry.Source, entry.Level, entry.Message, entry.CreatedAt)
	return err
}

func (r *LogRepository) ListBySource(ctx context.Context, source string, limit int) ([]LogEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, source, level, message, created_at FROM logs WHERE (? = '' OR source = ?) ORDER BY created_at DESC LIMIT ?`, source, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := []LogEntry{}
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.ID, &entry.Source, &entry.Level, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
