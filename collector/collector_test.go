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
	gaugeFunc := func(key string, val uint64) {
		keys = append(keys, key)
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
	// correctly. We expect there to be two of each key because there is one
	// from the initial stats, plus another one after the 1 second pause time.
	expCount := 2
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
	}
}
