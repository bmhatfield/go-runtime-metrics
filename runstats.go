package runstats

import "os"
import "fmt"
import "flag"
import "time"
import "strconv"
import "runtime"

import "github.com/bmhatfield/g2s"

var statsd *string = flag.String("statsd", "localhost:8125", "Statsd host:port pair")
var user_prefix *string = flag.String("metric-prefix", "default", "Metric prefix path; detects the local hostname by default")
var pause *int = flag.Int("pause", 10, "Collection pause interval")

var CPU *bool = flag.Bool("cpu", true, "Collect CPU Statistics")
var MEM *bool = flag.Bool("mem", true, "Collect Memory Statistics")
var GC *bool = flag.Bool("gc", true, "Collect GC Statistics (requires Memory be enabled)")

var prefix string
var s g2s.Statter

var err error

func init() {
	go collector()
}

func collector() {
	for !flag.Parsed() {
		// Defer execution of this goroutine.
		runtime.Gosched()
		
		// Add an initial delay while the program initializes to avoid attempting to collect
		// metrics prior to our flags being available / parsed.
		time.Sleep(time.Duration(1) * time.Second)
	}

	s, err = g2s.Dial("udp", *statsd)

	if err != nil {
		panic(fmt.Sprintf("Unable to connect to Statsd on %s - %s", *statsd, err))
	}

	if *user_prefix == "default" {
		prefix, err = os.Hostname()

		if err != nil {
			prefix = "go.unknown"
		} else {
			prefix = fmt.Sprintf("go.%s", prefix)
		}
	} else {
		prefix = *user_prefix
	}

	for {
		if *CPU {
			// Goroutines
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "cpu.goroutines"), strconv.Itoa(runtime.NumGoroutine()))

			// CGo calls
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "cpu.cgo_calls"), strconv.FormatUint(uint64(runtime.NumCgoCall()), 10))
		}

		if *MEM {
			m := new(runtime.MemStats)
			runtime.ReadMemStats(m)

			// General
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.alloc"), strconv.FormatUint(m.Alloc, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.total"), strconv.FormatUint(m.TotalAlloc, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.sys"), strconv.FormatUint(m.Sys, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.lookups"), strconv.FormatUint(m.Lookups, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.malloc"), strconv.FormatUint(m.Mallocs, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.frees"), strconv.FormatUint(m.Frees, 10))

			// Heap
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.alloc"), strconv.FormatUint(m.HeapAlloc, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.sys"), strconv.FormatUint(m.HeapSys, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.idle"), strconv.FormatUint(m.HeapIdle, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.inuse"), strconv.FormatUint(m.HeapInuse, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.released"), strconv.FormatUint(m.HeapReleased, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.heap.objects"), strconv.FormatUint(m.HeapObjects, 10))

			// Stack
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.inuse"), strconv.FormatUint(m.StackInuse, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.sys"), strconv.FormatUint(m.StackSys, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.mspan_inuse"), strconv.FormatUint(m.MSpanInuse, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.mspan_sys"), strconv.FormatUint(m.MSpanSys, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.mcache_inuse"), strconv.FormatUint(m.MCacheInuse, 10))
			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.stack.mcache_sys"), strconv.FormatUint(m.MCacheSys, 10))

			s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.othersys"), strconv.FormatUint(m.OtherSys, 10))

			if *GC {
				// GC
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.sys"), strconv.FormatUint(m.GCSys, 10))
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.next"), strconv.FormatUint(m.NextGC, 10))
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.last"), strconv.FormatUint(m.LastGC, 10))
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.pause_total"), strconv.FormatUint(m.PauseTotalNs, 10))
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.pause"), strconv.FormatUint(m.PauseNs[(m.NumGC+255)%256], 10))
				s.Gauge(1.0, fmt.Sprintf("%s.%s", prefix, "mem.gc.count"), strconv.FormatUint(uint64(m.NumGC), 10))
			}
		}

		// Gauges are a 'snapshot' rather than a histogram. Pausing for some interval
		// aims to get a 'recent' snapshot out before statsd flushes metrics.
		time.Sleep(time.Duration(*pause) * time.Second)
	}
}
