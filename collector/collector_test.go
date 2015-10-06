package collector

import (
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test because testing.Short is enabled")
	}

	var keys []string
	latestVal := make(map[string]uint64)
	gaugeFunc := func(key string, val uint64) {
		keys = append(keys, key)
		latestVal[key] = val
	}

	done := make(chan struct{})
	collectorShutdown := make(chan struct{})
	c := New(gaugeFunc)
	c.PauseDur = time.Second
	c.Done = done

	go func() {
		defer close(collectorShutdown)
		c.Run()
	}()
	time.Sleep(1500 * time.Millisecond)
	close(done)
	<-collectorShutdown

	// we're going to check a few keys to make sure that the stats are being output
	// correctly. We expect there to be three of each key because there is one
	// from the initial stats, plus another one after the 1 second pause time, plus
	// the final zeroing out on shutdown
	expCount := 3
	expKeys := []string{
		"cpu.goroutines",
		"mem.lookups",
		"mem.gc.count",
	}
	for _, expKey := range expKeys {
		count := 0
		for _, key := range keys {
			if key == expKey {
				count++
			}
		}
		if count != expCount {
			t.Errorf("unexpected num stats for key(%s):\ngot: %d\nexp: %d", expKey, count, expCount)
		}
		if latestVal[expKey] != 0 {
			t.Errorf("expected key (%s) to be zeroed out on shutdown, instead latest value is %d", expKey, latestVal[expKey])
		}
	}
}
