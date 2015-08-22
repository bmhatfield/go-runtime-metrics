package runstats

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/peterbourgon/g2s"
)

var (
	statsd *string = flag.String("statsd", "localhost:8125", "Statsd host:port pair")
	prefix *string = flag.String("metric-prefix", "<detect-hostname>", "Metric prefix path; detects the local hostname by default")
	pause  *int    = flag.Int("pause", 10, "Collection pause interval")

	cpu *bool = flag.Bool("cpu", true, "Collect CPU Statistics")
	mem *bool = flag.Bool("mem", true, "Collect Memory Statistics")
	gc  *bool = flag.Bool("gc", true, "Collect GC Statistics (requires Memory be enabled)")

	s g2s.Statter
)

func init() {
	go collector()
}

func collector() {
	for !flag.Parsed() {
		// Defer execution of this goroutine.
		runtime.Gosched()

		// Add an initial delay while the program initializes to avoid attempting to collect
		// metrics prior to our flags being available / parsed.
		time.Sleep(1 * time.Second)
	}

	var err error
	s, err = g2s.Dial("udp", *statsd)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to Statsd on %s - %s", *statsd, err))
	}

	if *prefix == "<detect-hostname>" {
		*prefix, err = os.Hostname()

		if err != nil {
			*prefix = "go.unknown"
		} else {
			*prefix = "go." + *prefix
		}
	}

	for {
		if *cpu {
			// Goroutines
			gaugeInt(s, "cpu.goroutines", runtime.NumGoroutine())

			// CGo calls
			gauge(s, "cpu.cgo_calls", uint64(runtime.NumCgoCall()))
		}

		if *mem {
			m := &runtime.MemStats{}
			runtime.ReadMemStats(m)

			// General
			gauge(s, "mem.alloc", m.Alloc)
			gauge(s, "mem.total", m.TotalAlloc)
			gauge(s, "mem.sys", m.Sys)
			gauge(s, "mem.lookups", m.Lookups)
			gauge(s, "mem.malloc", m.Mallocs)
			gauge(s, "mem.frees", m.Frees)

			// Heap
			gauge(s, "mem.heap.alloc", m.HeapAlloc)
			gauge(s, "mem.heap.sys", m.HeapSys)
			gauge(s, "mem.heap.idle", m.HeapIdle)
			gauge(s, "mem.heap.inuse", m.HeapInuse)
			gauge(s, "mem.heap.released", m.HeapReleased)
			gauge(s, "mem.heap.objects", m.HeapObjects)

			// Stack
			gauge(s, "mem.stack.inuse", m.StackInuse)
			gauge(s, "mem.stack.sys", m.StackSys)
			gauge(s, "mem.stack.mspan_inuse", m.MSpanInuse)
			gauge(s, "mem.stack.mspan_sys", m.MSpanSys)
			gauge(s, "mem.stack.mcache_inuse", m.MCacheInuse)
			gauge(s, "mem.stack.mcache_sys", m.MCacheSys)

			gauge(s, "mem.othersys", m.OtherSys)

			if *gc {
				// GC
				gauge(s, "mem.gc.sys", m.GCSys)
				gauge(s, "mem.gc.next", m.NextGC)
				gauge(s, "mem.gc.last", m.LastGC)
				gauge(s, "mem.gc.pause_total", m.PauseTotalNs)
				gauge(s, "mem.gc.pause", m.PauseNs[(m.NumGC+255)%256])
				gauge(s, "mem.gc.count", uint64(m.NumGC))
			}
		}

		// Gauges are a 'snapshot' rather than a histogram. Pausing for some interval
		// aims to get a 'recent' snapshot out before statsd flushes metrics.
		time.Sleep(time.Duration(*pause) * time.Second)
	}
}

func gaugeInt(s g2s.Statter, bucket string, val int) {
	s.Gauge(1, *prefix+"."+bucket, strconv.Itoa(val))
}

func gauge(s g2s.Statter, bucket string, val uint64) {
	s.Gauge(1, *prefix+"."+bucket, strconv.FormatUint(val, 10))
}
