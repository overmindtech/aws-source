package sources

import (
	"context"
	"testing"
	"time"
)

func TestMaxRefill(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := LimitBucket{
		MaxCapacity:    10,
		RefillRate:     7,
		RefillDuration: 10 * time.Millisecond,
		bucketFull:     make(chan bool),
	}

	b.Start(ctx)

	if full := <-b.bucketFull; full {
		t.Error("shouldn't be full on first refill")
	}

	if full := <-b.bucketFull; !full {
		t.Error("should be full on second refill")
	}
}

func TestWaiting(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := LimitBucket{
		MaxCapacity:    60,
		RefillRate:     20,
		RefillDuration: 100 * time.Millisecond,
		bucketFull:     make(chan bool),
	}

	b.Start(ctx)

	// Wait for the bucket to fill
	<-b.bucketFull
	<-b.bucketFull
	<-b.bucketFull

	// Delete the debug channel
	b.bucketFull = nil

	start := time.Now()

	// Execute 100 actions
	for i := 0; i < 100; i++ {
		// Get permission
		<-b.C
	}

	timeTaken := time.Since(start)

	// What should have happened is:
	//
	// 1. First 60 operations were near instant since the bucket was full
	// 2. Subsequent operations took place at a rate of 20 per duration
	if timeTaken < 100*time.Millisecond {
		t.Errorf("Should not have have been able to complete in <100ms, took %v", timeTaken.String())
	}

	if timeTaken > 500*time.Millisecond {
		t.Errorf("Should have have been able to complete in <500ms, took %v", timeTaken.String())
	}
}
