package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Lookups struct {
	pool *pgxpool.Pool
}

func NewLookups(pool *pgxpool.Pool) *Lookups {
	return &Lookups{pool: pool}
}

func (r *Lookups) ProgramIDByCode(ctx context.Context, code string) (string, bool, error) {
	const q = `SELECT id FROM programs WHERE code = $1 LIMIT 1`
	var id string
	err := r.pool.QueryRow(ctx, q, code).Scan(&id)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("program lookup: %w", err)
	}
	return id, true, nil
}

func (r *Lookups) LevelIDByName(ctx context.Context, name string) (string, bool, error) {
	const q = `SELECT id FROM levels WHERE name = $1 LIMIT 1`
	var id string
	err := r.pool.QueryRow(ctx, q, name).Scan(&id)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("level lookup: %w", err)
	}
	return id, true, nil
}

func (r *Lookups) GroupIDByName(ctx context.Context, name string) (string, bool, error) {
	const q = `SELECT id FROM groups WHERE name = $1 LIMIT 1`
	var id string
	err := r.pool.QueryRow(ctx, q, name).Scan(&id)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("group lookup: %w", err)
	}
	return id, true, nil
}

func (r *Lookups) SubjectIDsByCodes(ctx context.Context, codes []string) ([]string, error) {
	// Simple approach: query one by one for now; good enough for import sizes.
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		const q = `SELECT id FROM subjects WHERE code = $1 LIMIT 1`
		var id string
		err := r.pool.QueryRow(ctx, q, c).Scan(&id)
		if err != nil {
			if isNoRows(err) {
				continue
			}
			return nil, fmt.Errorf("subject lookup: %w", err)
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *Lookups) SubjectIDsByCodesStrict(ctx context.Context, codes []string) (ids []string, missing []string, err error) {
	ids = make([]string, 0, len(codes))
	missing = make([]string, 0)
	for _, c := range codes {
		const q = `SELECT id FROM subjects WHERE code = $1 LIMIT 1`
		var id string
		err := r.pool.QueryRow(ctx, q, c).Scan(&id)
		if err != nil {
			if isNoRows(err) {
				missing = append(missing, c)
				continue
			}
			return nil, nil, fmt.Errorf("subject lookup: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, missing, nil
}

func (r *Lookups) GroupIDsByNames(ctx context.Context, names []string) ([]string, error) {
	out := make([]string, 0, len(names))
	for _, n := range names {
		const q = `SELECT id FROM groups WHERE name = $1 LIMIT 1`
		var id string
		err := r.pool.QueryRow(ctx, q, n).Scan(&id)
		if err != nil {
			if isNoRows(err) {
				continue
			}
			return nil, fmt.Errorf("group lookup: %w", err)
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *Lookups) LevelIDsByNames(ctx context.Context, names []string) ([]string, error) {
	out := make([]string, 0, len(names))
	for _, n := range names {
		const q = `SELECT id FROM levels WHERE name = $1 LIMIT 1`
		var id string
		err := r.pool.QueryRow(ctx, q, n).Scan(&id)
		if err != nil {
			if isNoRows(err) {
				continue
			}
			return nil, fmt.Errorf("level lookup: %w", err)
		}
		out = append(out, id)
	}
	return out, nil
}
