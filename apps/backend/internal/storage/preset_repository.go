package storage

import (
	"context"
	"database/sql"
)

type PresetRepository struct {
	db *sql.DB
}

func NewPresetRepository(db *sql.DB) *PresetRepository {
	return &PresetRepository{db: db}
}

func (r *PresetRepository) List(ctx context.Context) ([]ZapretPreset, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, source, command, COALESCE(description, ''), is_active, updated_at FROM zapret_presets ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	presets := []ZapretPreset{}
	for rows.Next() {
		var p ZapretPreset
		var active int
		if err := rows.Scan(&p.ID, &p.Name, &p.Source, &p.Command, &p.Description, &active, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.IsActive = active == 1
		presets = append(presets, p)
	}
	return presets, rows.Err()
}

func (r *PresetRepository) Get(ctx context.Context, id string) (ZapretPreset, error) {
	var p ZapretPreset
	var active int
	err := r.db.QueryRowContext(ctx, `SELECT id, name, source, command, COALESCE(description, ''), is_active, updated_at FROM zapret_presets WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Source, &p.Command, &p.Description, &active, &p.UpdatedAt)
	if err != nil {
		return ZapretPreset{}, err
	}
	p.IsActive = active == 1
	return p, nil
}

func (r *PresetRepository) Upsert(ctx context.Context, p ZapretPreset) error {
	active := 0
	if p.IsActive {
		active = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO zapret_presets(id, name, source, command, description, is_active, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, source=excluded.source, command=excluded.command, description=excluded.description, is_active=excluded.is_active, updated_at=excluded.updated_at`,
		p.ID, p.Name, p.Source, p.Command, p.Description, active, p.UpdatedAt)
	return err
}

func (r *PresetRepository) MarkActive(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE zapret_presets SET is_active = 0`); err != nil {
		_ = tx.Rollback()
		return err
	}
	if id != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE zapret_presets SET is_active = 1 WHERE id = ?`, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
