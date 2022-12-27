package ec2

import (
	"context"
	"time"
)

// DefaultRefillDuration How often LimitBuckets are refilled by default
const DefaultRefillDuration = time.Second

// LimitBucket A struct that limits API usage in the same way that EC2 does:
// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/throttling.html
type LimitBucket struct {
	// The maximum number of tokens taht can be the bucket
	MaxCapacity int

	// How many tokens refill per refillDuration
	RefillRate int

	// How often tokens refill
	RefillDuration time.Duration

	// Channel tokens are stored in
	C <-chan struct{}
	c chan struct{} // Internal version of `C`

	// Channel that sends whicther or not the bucket is full each time the
	// bucket is refilled
	bucketFull chan bool
}

func (b *LimitBucket) Start(ctx context.Context) {
	if b.RefillDuration == 0 {
		b.RefillDuration = DefaultRefillDuration
	}

	tokenChan := make(chan struct{}, b.MaxCapacity)
	b.c = tokenChan
	b.C = tokenChan

	go func(ctx context.Context, bucket *LimitBucket) {
		ticker := time.NewTicker(bucket.RefillDuration)
		defer ticker.Stop()

		// Goroutine to continually refill
		for {
			select {
			case <-ticker.C:
				b.refill()
			case <-ctx.Done():
				return
			}
		}
	}(ctx, b)
}

// refill refuills the bucket the specified amount
func (b *LimitBucket) refill() {
	var newTokens int
	var full bool
	currentCapacity := len(b.c)

	// Make sure not to overfill the channel
	if delta := b.MaxCapacity - currentCapacity; delta < b.RefillRate {
		newTokens = delta
		full = true
	} else {
		newTokens = b.RefillRate
		full = false
	}

	// Refill the bucket
	for i := 0; i < newTokens; i++ {
		b.c <- struct{}{}
	}

	if b.bucketFull != nil {
		b.bucketFull <- full
	}
}
