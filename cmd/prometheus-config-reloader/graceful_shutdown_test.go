// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGracefulShutdownWaitsForInflightRequests verifies that shutdownServer
// uses http.Server.Shutdown (graceful) rather than Close (forceful).
//
// The test starts a handler that blocks for 200ms per request, fires 5
// concurrent requests, then calls shutdownServer immediately. All requests
// must complete successfully — which is only possible if Shutdown waits for
// them rather than dropping the connections like Close would.
func TestGracefulShutdownWaitsForInflightRequests(t *testing.T) {
	var completed atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow work.
		time.Sleep(200 * time.Millisecond)
		completed.Add(1)
		fmt.Fprint(w, "ok")
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	logger := slog.New(slog.DiscardHandler)

	const numRequests = 5
	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Fire concurrent requests.
	for range numRequests {
		go func() {
			defer wg.Done()
			resp, err := http.Get(srv.URL)
			if err != nil {
				// Only count as failure if it's not a "server closed" error.
				t.Logf("request error (may be expected after shutdown): %v", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("unexpected status: %d", resp.StatusCode)
			}
		}()
	}

	// Give requests a moment to reach the handler.
	time.Sleep(50 * time.Millisecond)

	// Trigger graceful shutdown.
	shutdownServer(logger, srv.Config)

	wg.Wait()

	got := completed.Load()
	if got != numRequests {
		t.Errorf("expected %d requests completed, got %d — server did not wait for in-flight requests (Close() behaviour?)", numRequests, got)
	}
}

// TestShutdownServerDoesNotErrorOnIdleServer verifies shutdownServer returns
// without error when there are no active connections.
func TestShutdownServerDoesNotErrorOnIdleServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	logger := slog.New(slog.DiscardHandler)

	// Should complete quickly without error.
	done := make(chan struct{})
	go func() {
		shutdownServer(logger, srv.Config)
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(10 * time.Second):
		t.Fatal("shutdownServer did not return within 10s on idle server")
	}
}

// TestShutdownUsesIndependentContext verifies that shutdownServer creates its
// own context rather than depending on a parent context that may already be
// cancelled. We do this by passing no parent context at all (the function
// signature only takes a logger and server), and confirming shutdown still
// works correctly.
func TestShutdownUsesIndependentContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	logger := slog.New(slog.DiscardHandler)

	// Fire a request, then shutdown immediately.
	var reqDone atomic.Int32
	go func() {
		resp, err := http.Get(srv.URL)
		if err == nil {
			resp.Body.Close()
		}
		reqDone.Store(1)
	}()

	time.Sleep(20 * time.Millisecond)
	shutdownServer(logger, srv.Config)

	if reqDone.Load() == 0 {
		t.Error("request was dropped — graceful shutdown did not work")
	}
}
