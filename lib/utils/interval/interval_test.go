// Copyright 2022 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interval

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

// TestLastTick verifies that the LastTick method returns the last tick time as expected.
func TestLastTick(t *testing.T) {
	clock := clockwork.NewFakeClock()
	interval := New(Config{
		Duration: time.Minute,
		Clock:    clock,
	})

	_, ok := interval.LastTick()
	require.False(t, ok)

	timeout := time.After(time.Second * 30)
	for i := 0; i < 3; i++ {
		clock.Advance(time.Minute)

		var tick time.Time
		select {
		case tick = <-interval.Next():
		case <-timeout:
			t.Fatal("timeout waiting for tick")
		}
		require.Equal(t, clock.Now(), tick)

		tick, ok = interval.LastTick()
		require.True(t, ok)
		require.Equal(t, clock.Now(), tick)
	}
}

// TestIntervalReset verifies the basic behavior of the interval reset functionality.
// Since time based tests tend to be flaky, this test passes if it has a >50% success
// rate (i.e. >50% of resets seemed to have actually extended the timer successfully).
func TestIntervalReset(t *testing.T) {
	const iterations = 1_000
	const duration = time.Millisecond * 666
	t.Parallel()

	var success, failure atomic.Uint64
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resetTimer := time.NewTimer(duration / 3)
			defer resetTimer.Stop()

			interval := New(Config{
				Duration: duration,
			})
			defer interval.Stop()

			start := time.Now()

			for i := 0; i < 6; i++ {
				select {
				case <-interval.Next():
					failure.Add(1)
					return
				case <-resetTimer.C:
					interval.Reset()
					resetTimer.Reset(duration / 3)
				}
			}

			<-interval.Next()
			elapsed := time.Since(start)
			// we expect this test to produce elapsed times of
			// 3*duration if it is working properly. we accept a
			// margin or error of +/- 1 duration in order to
			// minimize flakiness.
			if elapsed > duration*2 && elapsed < duration*4 {
				success.Add(1)
			} else {
				failure.Add(1)
			}
		}()
	}

	wg.Wait()

	require.Greater(t, success.Load(), failure.Load())
}

// TestIntervalResetTo verifies the basic behavior of the interval ResetTo method.
// Since time based tests tend to be flaky, this test passes if it has a >50% success
// rate (i.e. >50% of ResetTo calls seemed to have changed the timer's behavior as expected).
func TestIntervalResetTo(t *testing.T) {
	const workers = 1_000
	const ticks = 12
	const longDuration = time.Millisecond * 800
	const shortDuration = time.Millisecond * 200
	t.Parallel()

	var success, failure atomic.Uint64
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			interval := New(Config{
				Duration: longDuration,
			})
			defer interval.Stop()

			start := time.Now()

			for i := 0; i < ticks; i++ {
				interval.ResetTo(shortDuration)
				<-interval.Next()
			}

			elapsed := time.Since(start)
			// if the above works completed before the expected minimum time
			// to complete all ticks as long ticks, we assume that ResetTo has
			// successfully shortened the interval.
			if elapsed < longDuration*time.Duration(ticks) {
				success.Add(1)
			} else {
				failure.Add(1)
			}
		}()
	}

	wg.Wait()

	require.Greater(t, success.Load(), failure.Load())
}

func TestNewNoop(t *testing.T) {
	t.Parallel()
	i := NewNoop()
	ch := i.Next()
	select {
	case <-ch:
		t.Fatalf("noop should not emit anything")
	default:
	}
	i.Stop()
	select {
	case <-ch:
		t.Fatalf("noop should not emit anything")
	default:
	}
}
