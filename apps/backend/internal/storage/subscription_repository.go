package storage

import (
	"context"
	"database/sql"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, url, update_interval, COALESCE(last_update_at, ''), created_at FROM subscriptions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Subscription{}
	for rows.Next() {
		var item Subscription
		if err := rows.Scan(&item.ID, &item.Name, &item.URL, &item.UpdateInterval, &item.LastUpdateAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SubscriptionRepository) Get(ctx context.Context, id string) (Subscription, error) {
	var item Subscription
	err := r.db.QueryRowContext(ctx, `SELECT id, name, url, update_interval, COALESCE(last_update_at, ''), created_at FROM subscriptions WHERE id = ?`, id).
		Scan(&item.ID, &item.Name, &item.URL, &item.UpdateInterval, &item.LastUpdateAt, &item.CreatedAt)
	if err != nil {
		return Subscription{}, err
	}
	return item, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, item Subscription) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO subscriptions(id, name, url, update_interval, last_update_at, created_at) VALUES (?, ?, ?, ?, NULLIF(?, ''), ?)`,
		item.ID, item.Name, item.URL, item.UpdateInterval, item.LastUpdateAt, item.CreatedAt)
	return err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = ?`, id)
	return err
}

func (r *SubscriptionRepository) MarkUpdated(ctx context.Context, id string, updatedAt string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET last_update_at = ? WHERE id = ?`, updatedAt, id)
	return err
}
