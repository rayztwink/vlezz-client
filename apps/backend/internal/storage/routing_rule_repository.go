package storage

import (
	"context"
	"database/sql"
)

type RoutingRuleRepository struct {
	db *sql.DB
}

func NewRoutingRuleRepository(db *sql.DB) *RoutingRuleRepository {
	return &RoutingRuleRepository{db: db}
}

func (r *RoutingRuleRepository) List(ctx context.Context) ([]RoutingRule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, domain, mode, enabled FROM routing_rules ORDER BY domain ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := []RoutingRule{}
	for rows.Next() {
		var rule RoutingRule
		var enabled int
		if err := rows.Scan(&rule.ID, &rule.Domain, &rule.Mode, &enabled); err != nil {
			return nil, err
		}
		rule.Enabled = enabled == 1
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *RoutingRuleRepository) Create(ctx context.Context, rule RoutingRule) error {
	enabled := 0
	if rule.Enabled {
		enabled = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO routing_rules(id, domain, mode, enabled) VALUES (?, ?, ?, ?)`, rule.ID, rule.Domain, rule.Mode, enabled)
	return err
}

func (r *RoutingRuleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM routing_rules WHERE id = ?`, id)
	return err
}
