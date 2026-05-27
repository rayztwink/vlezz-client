package storage

import (
	"context"
	"database/sql"
)

type NodeRepository struct {
	db *sql.DB
}

func NewNodeRepository(db *sql.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) List(ctx context.Context) ([]Node, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, protocol, address, port, uuid, security, transport, raw_link, latency_ms, country, created_at FROM nodes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		var n Node
		var rawLink sql.NullString
		var latency sql.NullInt64
		var country sql.NullString
		if err := rows.Scan(&n.ID, &n.Name, &n.Protocol, &n.Address, &n.Port, &n.UUID, &n.Security, &n.Transport, &rawLink, &latency, &country, &n.CreatedAt); err != nil {
			return nil, err
		}
		if rawLink.Valid {
			n.RawLink = rawLink.String
		}
		if latency.Valid {
			value := int(latency.Int64)
			n.LatencyMS = &value
		}
		if country.Valid {
			n.Country = country.String
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (r *NodeRepository) Get(ctx context.Context, id string) (Node, error) {
	var n Node
	var rawLink sql.NullString
	var latency sql.NullInt64
	var country sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, name, protocol, address, port, uuid, security, transport, raw_link, latency_ms, country, created_at FROM nodes WHERE id = ?`, id).
		Scan(&n.ID, &n.Name, &n.Protocol, &n.Address, &n.Port, &n.UUID, &n.Security, &n.Transport, &rawLink, &latency, &country, &n.CreatedAt)
	if err != nil {
		return Node{}, err
	}
	if rawLink.Valid {
		n.RawLink = rawLink.String
	}
	if latency.Valid {
		value := int(latency.Int64)
		n.LatencyMS = &value
	}
	if country.Valid {
		n.Country = country.String
	}
	return n, nil
}

func (r *NodeRepository) Create(ctx context.Context, n Node) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO nodes(id, name, protocol, address, port, uuid, security, transport, raw_link, latency_ms, country, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.Name, n.Protocol, n.Address, n.Port, n.UUID, n.Security, n.Transport, n.RawLink, n.LatencyMS, n.Country, n.CreatedAt)
	return err
}

func (r *NodeRepository) ExistsRawLink(ctx context.Context, rawLink string) (bool, error) {
	if rawLink == "" {
		return false, nil
	}
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM nodes WHERE raw_link = ?`, rawLink).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *NodeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	return err
}

func (r *NodeRepository) UpdateLatency(ctx context.Context, id string, latencyMs int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE nodes SET latency_ms = ? WHERE id = ?`, latencyMs, id)
	return err
}
