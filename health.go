package core

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"
)

// Kind represents the type of health check (liveness or readiness).
type Kind string

const (
	// Liveness indicates the service is alive.
	Liveness Kind = "liveness"
	// Readiness indicates the service is ready to serve traffic.
	Readiness Kind = "readiness"
)

// Check defines a health check that can be registered.
type Check interface {
	Name() string
	Kind() Kind
	Check(ctx context.Context) error
}

// Result represents the aggregated result of health checks.
type Result struct {
	OK      bool
	Details map[string]struct {
		OK    bool
		Error string
	}
}

// Registry manages health checks and their status.
type Registry interface {
	Register(c Check)
	Aggregate(ctx context.Context, kind Kind) Result
	Set(kind Kind, name string, err error)
}

type healthRegistry struct {
	mu     sync.RWMutex
	checks map[Kind]map[string]Check
	status map[Kind]map[string]error
}

// NewHealthRegistry creates a new health check registry.
func NewHealthRegistry() Registry {
	return &healthRegistry{
		checks: make(map[Kind]map[string]Check),
		status: make(map[Kind]map[string]error),
	}
}

func (r *healthRegistry) Register(c Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.checks[c.Kind()] == nil {
		r.checks[c.Kind()] = make(map[string]Check)
	}
	r.checks[c.Kind()][c.Name()] = c
}

func (r *healthRegistry) Aggregate(ctx context.Context, kind Kind) Result {
	r.mu.RLock()
	checks := r.checks[kind]
	r.mu.RUnlock()

	res := Result{Details: make(map[string]struct {
		OK    bool
		Error string
	})}
	timeout := 300 * time.Millisecond
	if tms, ok := os.LookupEnv("STRATUM_HEALTH_TIMEOUT_MS"); ok {
		if ms, _ := strconv.Atoi(tms); ms > 0 {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	res.OK = true
	for name, c := range checks {
		wg.Add(1)
		go func(name string, c Check) {
			defer wg.Done()
			ctxCheck, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			err := c.Check(ctxCheck)
			mu.Lock()
			if err != nil {
				res.OK = false
				res.Details[name] = struct {
					OK    bool
					Error string
				}{false, err.Error()}
			} else {
				res.Details[name] = struct {
					OK    bool
					Error string
				}{true, ""}
			}
			mu.Unlock()
		}(name, c)
	}
	wg.Wait()
	return res
}

func (r *healthRegistry) Set(kind Kind, name string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.status[kind] == nil {
		r.status[kind] = make(map[string]error)
	}
	r.status[kind][name] = err
}
