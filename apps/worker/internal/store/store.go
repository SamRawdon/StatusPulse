package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

type DueMonitor struct {
	ID              string
	URL             string
	Method          string
	ExpectedStatus  int
	IntervalSeconds int
	TimeoutMs       int
}

type CheckRunIn struct {
	MonitorID  string
	CheckedAt  time.Time
	StatusCode *int
	LatencyMs  *int
	Success    bool
	ErrorText  *string
}

func (s *Store) ListDueMonitors(ctx context.Context, limit int) ([]DueMonitor, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.Query(ctx, `
		select id, url, method, expected_status, interval_seconds, timeout_ms
		from monitors
		where enabled = true
		  and next_run_at <= now()
		order by next_run_at asc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DueMonitor
	for rows.Next() {
		var m DueMonitor
		if err := rows.Scan(&m.ID, &m.URL, &m.Method, &m.ExpectedStatus, &m.IntervalSeconds, &m.TimeoutMs); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) InsertCheckRun(ctx context.Context, in CheckRunIn) error {
	_, err := s.DB.Exec(ctx, `
		insert into check_runs (monitor_id, checked_at, status_code, latency_ms, success, error_text)
		values ($1, $2, $3, $4, $5, $6)
	`, in.MonitorID, in.CheckedAt, in.StatusCode, in.LatencyMs, in.Success, in.ErrorText)
	return err
}

func (s *Store) BumpNextRun(ctx context.Context, monitorID string, intervalSeconds int) error {
	_, err := s.DB.Exec(ctx, `
		update monitors
		set next_run_at = now() + ($2::text || ' seconds')::interval
		where id = $1
	`, monitorID, intervalSeconds)
	return err
}
