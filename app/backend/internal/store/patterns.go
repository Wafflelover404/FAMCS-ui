package store

import (
	"context"
	"encoding/json"
	"time"
)

type PatternRecord struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	N         int       `json:"n"`
	K         int       `json:"k"`
	Cells     []int     `json:"cells"`
	Stock     []int     `json:"stock,omitempty"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Store) SavePattern(ctx context.Context, slug, title string, n, k int, cells, stock []int) (string, error) {
	cellsJSON, err := json.Marshal(cells)
	if err != nil {
		return "", err
	}
	var stockJSON []byte
	if stock != nil {
		stockJSON, err = json.Marshal(stock)
		if err != nil {
			return "", err
		}
	}
	var id string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO patterns (slug, n, k, cells, stock, title) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (slug) DO UPDATE SET n=$2, k=$3, cells=$4, stock=$5, title=$6
		RETURNING id`, slug, n, k, cellsJSON, stockJSON, title,
	).Scan(&id)
	return id, err
}

func (s *Store) GetPattern(ctx context.Context, slug string) (*PatternRecord, error) {
	var r PatternRecord
	var cellsJSON, stockJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, slug, n, k, cells, stock, title, created_at FROM patterns WHERE slug=$1`, slug,
	).Scan(&r.ID, &r.Slug, &r.N, &r.K, &cellsJSON, &stockJSON, &r.Title, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cellsJSON, &r.Cells); err != nil {
		return nil, err
	}
	if stockJSON != nil {
		if err := json.Unmarshal(stockJSON, &r.Stock); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

func (s *Store) ListPatterns(ctx context.Context) ([]PatternRecord, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, slug, n, k, title, created_at FROM patterns ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PatternRecord
	for rows.Next() {
		var r PatternRecord
		if err := rows.Scan(&r.ID, &r.Slug, &r.N, &r.K, &r.Title, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
