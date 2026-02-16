package checker

import (
	"context"
	"net/http"
	"time"
)

type Result struct {
	StatusCode *int
	LatencyMs  *int
	Success    bool
	ErrorText  *string
}

func Run(ctx context.Context, method, url string, expectedStatus int, timeoutMs int) Result {
	client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		msg := err.Error()
		return Result{Success: false, ErrorText: &msg}
	}

	start := time.Now()
	resp, err := client.Do(req)
	lat := int(time.Since(start).Milliseconds())

	if err != nil {
		msg := err.Error()
		return Result{LatencyMs: &lat, Success: false, ErrorText: &msg}
	}
	defer resp.Body.Close()

	sc := resp.StatusCode
	ok := sc == expectedStatus

	return Result{StatusCode: &sc, LatencyMs: &lat, Success: ok}
}
