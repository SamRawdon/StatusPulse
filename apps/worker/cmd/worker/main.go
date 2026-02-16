package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"statuspulse/apps/worker/internal/checker"
	"statuspulse/apps/worker/internal/store"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	s := &store.Store{DB: db}

	log.Println("worker starting")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return
		case <-ticker.C:
			runTick(ctx, s)
		}
	}
}

func runTick(ctx context.Context, s *store.Store) {
	ms, err := s.ListDueMonitors(ctx, 200)
	if err != nil {
		log.Println("ListDueMonitors:", err)
		return
	}

	for _, m := range ms {
		res := checker.Run(ctx, m.Method, m.URL, m.ExpectedStatus, m.TimeoutMs)

		now := time.Now()

		cr := store.CheckRunIn{
			MonitorID:  m.ID,
			CheckedAt:  now,
			StatusCode: res.StatusCode,
			LatencyMs:  res.LatencyMs,
			Success:    res.Success,
			ErrorText:  res.ErrorText,
		}

		if err := s.InsertCheckRun(ctx, cr); err != nil {
			log.Println("InsertCheckRun:", err)
			continue
		}

		if err := s.BumpNextRun(ctx, m.ID, m.IntervalSeconds); err != nil {
			log.Println("BumpNextRun:", err)
			continue
		}

		if res.Success {
			log.Printf("check ok monitor=%s latency_ms=%v", m.ID, valOrNil(res.LatencyMs))
		} else {
			log.Printf("check fail monitor=%s err=%v status=%v", m.ID, strOrNil(res.ErrorText), valOrNil(res.StatusCode))
		}
	}
}

func valOrNil[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}

func strOrNil(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}
