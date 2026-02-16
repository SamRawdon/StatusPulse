package httpapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"statuspulse/apps/api/internal/store"
)

type API struct {
	Store *store.Store
}

func Router(a *API, corsOrigin string) http.Handler {
	r := chi.NewRouter()

	allowed := []string{"http://localhost:3000"}
	if corsOrigin != "" {
		allowed = append(allowed, corsOrigin)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowed,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/monitors", a.listMonitors)
		r.Post("/monitors", a.createMonitor)
		r.Get("/monitors/{id}/checks", a.listChecksForMonitor)
	})

	return r
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (a *API) listMonitors(w http.ResponseWriter, r *http.Request) {
	ms, err := a.Store.ListMonitors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, ms)
}

func (a *API) createMonitor(w http.ResponseWriter, r *http.Request) {
	var in store.CreateMonitorInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if in.Method == "" {
		in.Method = "GET"
	}
	if in.ExpectedStatus == 0 {
		in.ExpectedStatus = 200
	}
	if in.IntervalSeconds == 0 {
		in.IntervalSeconds = 60
	}
	if in.TimeoutMs == 0 {
		in.TimeoutMs = 5000
	}

	if in.Name == "" {
		http.Error(w, "name is required", 400)
		return
	}
	parsed, err := url.Parse(in.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		http.Error(w, "url must be a valid absolute URL", 400)
		return
	}
	if in.IntervalSeconds < 10 || in.IntervalSeconds > 3600 {
		http.Error(w, "intervalSeconds must be between 10 and 3600", 400)
		return
	}
	if in.TimeoutMs < 100 || in.TimeoutMs > 60000 {
		http.Error(w, "timeoutMs must be between 100 and 60000", 400)
		return
	}

	out, err := a.Store.CreateMonitor(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 201, out)
}

func (a *API) listChecksForMonitor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	out, err := a.Store.ListChecksForMonitor(r.Context(), id, 50)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, out)
}
