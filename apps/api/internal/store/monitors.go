package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

type Monitor struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Method          string    `json:"method"`
	ExpectedStatus  int       `json:"expectedStatus"`
	IntervalSeconds int       `json:"intervalSeconds"`
	TimeoutMs       int       `json:"timeoutMs"`
	Enabled         bool      `json:"enabled"`
	NextRunAt       time.Time `json:"nextRunAt"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CheckRun struct {
	ID        string     `json:"id"`
	MonitorID string     `json:"monitorId"`
	CheckedAt time.Time  `json:"checkedAt"`
	Status    *int       `json:"statusCode"`
	LatencyMs *int       `json:"latencyMs"`
	Success   bool       `json:"success"`
	ErrorText *string    `json:"errorText"`
}

type CreateMonitorInput struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Method          string `json:"method"`
	ExpectedStatus  int    `json:"expectedStatus"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutMs       int    `json:"timeoutMs"`
}

func (s *Store) ListMonitors(ctx context.Context) ([]Monitor, error) {
	rows, err := s.DB.Query(ctx, `
		select id, name, url, method, expected_status, interval_seconds, timeout_ms, enabled, next_run_at, created_at
		from monitors
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Monitor
	for rows.Next() {
		var m Monitor
		if err := rows.Scan(
			&m.ID, &m.Name, &m.URL, &m.Method, &m.ExpectedStatus, &m.IntervalSeconds, &m.TimeoutMs,
			&m.Enabled, &m.NextRunAt, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) CreateMonitor(ctx context.Context, in CreateMonitorInput) (Monitor, error) {
	var m Monitor
	err := s.DB.QueryRow(ctx, `
		insert into monitors (name, url, method, expected_status, interval_seconds, timeout_ms, enabled, next_run_at)
		values ($1, $2, $3, $4, $5, $6, true, now())
		returning id, name, url, method, expected_status, interval_seconds, timeout_ms, enabled, next_run_at, created_at
	`, in.Name, in.URL, in.Method, in.ExpectedStatus, in.IntervalSeconds, in.TimeoutMs).Scan(
		&m.ID, &m.Name, &m.URL, &m.Method, &m.ExpectedStatus, &m.IntervalSeconds, &m.TimeoutMs,
		&m.Enabled, &m.NextRunAt, &m.CreatedAt,
	)
	return m, err
}

func (s *Store) ListChecksForMonitor(ctx context.Context, monitorID string, limit int) ([]CheckRun, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.Query(ctx, `
		select id, monitor_id, checked_at, status_code, latency_ms, success, error_text
		from check_runs
		where monitor_id = $1
		order by checked_at desc
		limit $2
	`, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CheckRun
	for rows.Next() {
		var r CheckRun
		if err := rows.Scan(
			&r.ID, &r.MonitorID, &r.CheckedAt, &r.Status, &r.LatencyMs, &r.Success, &r.ErrorText,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
