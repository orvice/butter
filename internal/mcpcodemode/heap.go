package mcpcodemode

import (
	"context"
	"runtime/metrics"
	"time"
)

const heapPollInterval = 25 * time.Millisecond

// liveHeapMetric is the bytes of live and not-yet-swept heap objects.
const liveHeapMetric = "/memory/classes/heap/objects:bytes"

func heapObjectBytes() uint64 {
	sample := []metrics.Sample{{Name: liveHeapMetric}}
	metrics.Read(sample)
	if sample[0].Value.Kind() != metrics.KindUint64 {
		return 0
	}
	return sample[0].Value.Uint64()
}

// heapWatch calls stop(growth) once the process's heap has grown by more than
// limit bytes over what it held when the watch started, and then stops
// watching. It ends with ctx.
//
// goja cannot cap a script's memory, so this is a backstop for scripts that
// grow their heap over time (a loop filling an array, a string doubled in a
// loop); it is process-wide, which is why the limit is generous. It cannot
// stop one allocation that is itself larger than the limit, nor a script that
// outruns its 25ms poll.
func heapWatch(ctx context.Context, limit uint64, stop func(growth uint64)) {
	base := heapObjectBytes()
	tick := time.NewTicker(heapPollInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if now := heapObjectBytes(); now > base && now-base > limit {
				stop(now - base)
				return
			}
		}
	}
}
