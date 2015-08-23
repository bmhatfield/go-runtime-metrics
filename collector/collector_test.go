package collector

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test because testing.Short is enabled")
	}

	// create a UDP listener so that our Collector can write to it
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to start up upd listener: %v", err)
	}
	defer conn.Close()
	done := make(chan struct{})
	updShutdown := make(chan struct{})
	collectorShutdown := make(chan struct{})
	config := DefaultConfig
	config.Addr = conn.LocalAddr().String()
	config.PauseDur = time.Second
	config.Done = done
	c, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	// spin off a goroutine to collect all the statsd messages written by Collector
	var stats [][]byte
	go func() {
		defer close(updShutdown)
		b := make([]byte, 1024*1204)
		for {
			select {
			case <-done:
				return
			default:
				n, _, err := conn.ReadFromUDP(b)
				if err != nil {
					continue
				}
				stats = append(stats, append([]byte{}, b[:n]...))
			}
		}
	}()
	go func() {
		defer close(collectorShutdown)
		c.Run()
	}()
	time.Sleep(1500 * time.Millisecond)
	close(done)
	conn.Close()
	<-updShutdown
	<-collectorShutdown

	// we're going to check a few keys to make sure that the stats are being output
	// correctly. We expect there to be two of each key because there is one
	// from the initial stats, plus another one after the 1 second pause time.
	expCount := 2
	expKeys := [][]byte{
		[]byte("cpu.goroutines"),
		[]byte("mem.lookups"),
		[]byte("mem.gc.count"),
	}
	for _, key := range expKeys {
		count := 0
		for _, stat := range stats {
			count += bytes.Count(stat, key)
		}
		if count != expCount {
			t.Errorf("unexpected num stats for key(%s):\ngot: %d\nexp: %d", key, count, expCount)
		}
	}
}
