package storage

import (
	"context"
	"database/sql"
)

type CheckRepository struct {
	db *sql.DB
}

func NewCheckRepository(db *sql.DB) *CheckRepository {
	return &CheckRepository{db: db}
}

func (r *CheckRepository) Create(ctx context.Context, c DiagnosticCheck) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO checks(id, target, mode, status, latency_ms, error, checked_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Target, c.Mode, c.Status, c.LatencyMS, c.Error, c.CheckedAt)
	return err
}

func (r *CheckRepository) History(ctx context.Context, limit int) ([]DiagnosticCheck, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, target, mode, status, latency_ms, COALESCE(error, ''), checked_at FROM checks ORDER BY checked_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	checks := []DiagnosticCheck{}
	for rows.Next() {
		var c DiagnosticCheck
		var latency sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Target, &c.Mode, &c.Status, &latency, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		if latency.Valid {
			value := int(latency.Int64)
			c.LatencyMS = &value
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}
