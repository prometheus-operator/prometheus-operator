// Copyright 2024 The prometheus-operator Authors
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

package operator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

// mockSyncer implements the Syncer interface for testing
type mockSyncer struct {
	mu               sync.Mutex
	syncFunc         func(context.Context, string) error
	updateStatusFunc func(context.Context, string) error
	syncCalls        []string
	statusCalls      []string
}

func (m *mockSyncer) Sync(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncCalls = append(m.syncCalls, key)
	if m.syncFunc != nil {
		return m.syncFunc(ctx, key)
	}
	return nil
}

func (m *mockSyncer) UpdateStatus(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCalls = append(m.statusCalls, key)
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, key)
	}
	return nil
}

func (m *mockSyncer) getSyncCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.syncCalls)
}

// mockGetter implements the OwnedResourceOwner interface for testing
type mockGetter struct {
	getFunc func(string) (runtime.Object, error)
}

func (m *mockGetter) Get(key string) (runtime.Object, error) {
	if m.getFunc != nil {
		return m.getFunc(key)
	}
	return nil, nil
}

// mockReconcilerMetrics implements the ReconcilerMetrics interface for testing
type mockReconcilerMetrics struct{}

func (m *mockReconcilerMetrics) TriggerByCounter(string, HandlerEvent) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{Name: "test_counter"})
}

func newTestResourceReconciler(t *testing.T) (*ResourceReconciler, *mockSyncer) {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	syncer := &mockSyncer{}
	getter := &mockGetter{}
	metrics := &mockReconcilerMetrics{}
	reg := prometheus.NewRegistry()

	rr := NewResourceReconciler(
		logger,
		syncer,
		getter,
		metrics,
		"TestResource",
		reg,
		"test-controller",
	)

	return rr, syncer
}

func TestSetReconcileDelay(t *testing.T) {
	rr, _ := newTestResourceReconciler(t)

	tests := []struct {
		name          string
		delay         time.Duration
		expectedDelay time.Duration
	}{
		{
			name:          "zero delay",
			delay:         0,
			expectedDelay: 0,
		},
		{
			name:          "30 second delay",
			delay:         30 * time.Second,
			expectedDelay: 30 * time.Second,
		},
		{
			name:          "5 minute delay",
			delay:         5 * time.Minute,
			expectedDelay: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr.SetReconcileDelay(tt.delay)
			assert.Equal(t, tt.expectedDelay, rr.reconcileDelay)
		})
	}
}

func TestShouldDelayReconcile(t *testing.T) {
	rr, _ := newTestResourceReconciler(t)

	tests := []struct {
		name           string
		delay          time.Duration
		key            string
		lastReconcile  *time.Time
		expectedResult bool
	}{
		{
			name:           "no delay configured",
			delay:          0,
			key:            "test/resource",
			lastReconcile:  nil,
			expectedResult: false,
		},
		{
			name:           "delay configured but no previous reconciliation",
			delay:          30 * time.Second,
			key:            "test/resource",
			lastReconcile:  nil,
			expectedResult: false,
		},
		{
			name:           "delay configured and within delay period",
			delay:          30 * time.Second,
			key:            "test/resource",
			lastReconcile:  timePtr(time.Now().Add(-10 * time.Second)),
			expectedResult: true,
		},
		{
			name:           "delay configured and delay period has passed",
			delay:          30 * time.Second,
			key:            "test/resource",
			lastReconcile:  timePtr(time.Now().Add(-35 * time.Second)),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the reconciler state
			rr.SetReconcileDelay(tt.delay)
			rr.lastReconcile = make(map[string]time.Time)

			// Set up last reconcile time if provided
			if tt.lastReconcile != nil {
				rr.lastReconcile[tt.key] = *tt.lastReconcile
			}

			result := rr.shouldDelayReconcile(tt.key)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestUpdateLastReconcileTime(t *testing.T) {
	rr, _ := newTestResourceReconciler(t)

	key := "test/resource"
	beforeUpdate := time.Now()

	rr.updateLastReconcileTime(key)

	afterUpdate := time.Now()
	recordedTime, exists := rr.lastReconcile[key]

	assert.True(t, exists, "Last reconcile time should be recorded")
	assert.True(t, recordedTime.After(beforeUpdate) || recordedTime.Equal(beforeUpdate))
	assert.True(t, recordedTime.Before(afterUpdate) || recordedTime.Equal(afterUpdate))
}

func TestProcessNextReconcileItemWithDelay(t *testing.T) {
	rr, syncer := newTestResourceReconciler(t)

	// Set up a 100ms delay
	delay := 100 * time.Millisecond
	rr.SetReconcileDelay(delay)

	key := "test/resource"

	// Add item to the queue
	rr.reconcileQ.Add(key)

	// Process the item for the first time (should succeed immediately)
	ctx := context.Background()
	result := rr.processNextReconcileItem(ctx)

	assert.True(t, result, "First reconciliation should succeed")
	assert.Equal(t, 1, syncer.getSyncCallCount(), "Sync should be called once")

	// Add the same item again immediately
	rr.reconcileQ.Add(key)

	// Process the item again (should be delayed)
	result = rr.processNextReconcileItem(ctx)

	assert.True(t, result, "Processing should return true even when delayed")
	assert.Equal(t, 1, syncer.getSyncCallCount(), "Sync should not be called again due to delay")

	// Wait for the delay period and try again
	time.Sleep(delay + 10*time.Millisecond)

	// Add item again and process
	rr.reconcileQ.Add(key)
	result = rr.processNextReconcileItem(ctx)

	assert.True(t, result, "Processing should succeed after delay period")
	assert.Equal(t, 2, syncer.getSyncCallCount(), "Sync should be called again after delay")
}

func TestProcessNextReconcileItemWithoutDelay(t *testing.T) {
	rr, syncer := newTestResourceReconciler(t)

	// No delay configured (default is 0)
	key := "test/resource"

	// Add and process item multiple times
	for i := 0; i < 3; i++ {
		rr.reconcileQ.Add(key)
		ctx := context.Background()
		result := rr.processNextReconcileItem(ctx)

		assert.True(t, result, fmt.Sprintf("Processing attempt %d should succeed", i+1))
		assert.Equal(t, i+1, syncer.getSyncCallCount(), fmt.Sprintf("Sync should be called %d times", i+1))
	}
}

func TestProcessNextReconcileItemWithSyncError(t *testing.T) {
	rr, syncer := newTestResourceReconciler(t)

	// Set up syncer to return an error
	expectedError := fmt.Errorf("sync failed")
	syncer.syncFunc = func(ctx context.Context, key string) error {
		return expectedError
	}

	key := "test/resource"
	rr.reconcileQ.Add(key)

	ctx := context.Background()
	result := rr.processNextReconcileItem(ctx)

	assert.True(t, result, "Processing should return true even on sync error")
	assert.Equal(t, 1, syncer.getSyncCallCount(), "Sync should be called once")

	// Last reconcile time should NOT be updated on error
	_, exists := rr.lastReconcile[key]
	assert.False(t, exists, "Last reconcile time should not be updated on sync error")
}

func TestDelayIntegrationWithWorkQueue(t *testing.T) {
	rr, syncer := newTestResourceReconciler(t)

	delay := 50 * time.Millisecond
	rr.SetReconcileDelay(delay)

	key := "test/resource"

	// Start the reconciler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go rr.Run(ctx)
	defer rr.Stop()

	// Add item to queue
	rr.reconcileQ.Add(key)

	// Wait a bit for first processing
	assert.Eventually(t, func() bool {
		return syncer.getSyncCallCount() == 1
	}, 100*time.Millisecond, 10*time.Millisecond, "First sync should happen immediately")

	// Add same item again (should be delayed)
	rr.reconcileQ.Add(key)

	// Wait shorter than delay period
	time.Sleep(delay / 2)
	assert.Equal(t, 1, syncer.getSyncCallCount(), "Second sync should be delayed")

	// Wait for delay period to complete
	time.Sleep(delay + 20*time.Millisecond)
	assert.Eventually(t, func() bool {
		return syncer.getSyncCallCount() == 2
	}, 100*time.Millisecond, 10*time.Millisecond, "Second sync should happen after delay")
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
