package collector

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/peterbourgon/g2s"
)

type Config struct {
	Addr      string
	Prefix    string
	PauseDur  time.Duration
	EnableCPU bool
	EnableMem bool
	EnableGC  bool
	Done      <-chan struct{}
}

var DefaultConfig = Config{
	Addr:      "localhost:8125",
	Prefix:    "<detect-hostname>",
	PauseDur:  10 * time.Second,
	EnableCPU: true,
	EnableMem: true,
	EnableGC:  true,
}

type Collector struct {
	config Config
	s      g2s.Statter
}

func New(config Config) (*Collector, error) {
	s, err := g2s.Dial("udp", config.Addr)
	if err != nil {
		return nil, err
	}
	if config.Prefix == "<detect-hostname>" {
		hn, err := os.Hostname()

		if err != nil {
			config.Prefix = "go.unknown"
		} else {
			config.Prefix = "go." + hn
		}
	}
	return &Collector{config: config, s: s}, nil
}

func (c *Collector) Run() {
	c.outputStats()

	// Gauges are a 'snapshot' rather than a histogram. Pausing for some interval
	// aims to get a 'recent' snapshot out before statsd flushes metrics.
	tick := time.NewTicker(c.config.PauseDur)
	defer tick.Stop()
	for {
		select {
		case <-c.config.Done:
			return
		case <-tick.C:
			c.outputStats()
		}
	}
}

func (c *Collector) outputStats() {
	if c.config.EnableCPU {
		// Goroutines
		c.gaugeInt("cpu.goroutines", runtime.NumGoroutine())

		// CGo calls
		c.gauge("cpu.cgo_calls", uint64(runtime.NumCgoCall()))
	}

	if c.config.EnableMem {
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)

		// General
		c.gauge("mem.alloc", m.Alloc)
		c.gauge("mem.total", m.TotalAlloc)
		c.gauge("mem.sys", m.Sys)
		c.gauge("mem.lookups", m.Lookups)
		c.gauge("mem.malloc", m.Mallocs)
		c.gauge("mem.frees", m.Frees)

		// Heap
		c.gauge("mem.heap.alloc", m.HeapAlloc)
		c.gauge("mem.heap.sys", m.HeapSys)
		c.gauge("mem.heap.idle", m.HeapIdle)
		c.gauge("mem.heap.inuse", m.HeapInuse)
		c.gauge("mem.heap.released", m.HeapReleased)
		c.gauge("mem.heap.objects", m.HeapObjects)

		// Stack
		c.gauge("mem.stack.inuse", m.StackInuse)
		c.gauge("mem.stack.sys", m.StackSys)
		c.gauge("mem.stack.mspan_inuse", m.MSpanInuse)
		c.gauge("mem.stack.mspan_sys", m.MSpanSys)
		c.gauge("mem.stack.mcache_inuse", m.MCacheInuse)
		c.gauge("mem.stack.mcache_sys", m.MCacheSys)

		c.gauge("mem.othersys", m.OtherSys)

		if c.config.EnableGC {
			// GC
			c.gauge("mem.gc.sys", m.GCSys)
			c.gauge("mem.gc.next", m.NextGC)
			c.gauge("mem.gc.last", m.LastGC)
			c.gauge("mem.gc.pause_total", m.PauseTotalNs)
			c.gauge("mem.gc.pause", m.PauseNs[(m.NumGC+255)%256])
			c.gauge("mem.gc.count", uint64(m.NumGC))
		}
	}
}

func (c *Collector) gaugeInt(bucket string, val int) {
	c.s.Gauge(1.0, c.config.Prefix+"."+bucket, strconv.Itoa(val))
}

func (c *Collector) gauge(bucket string, val uint64) {
	c.s.Gauge(1.0, c.config.Prefix+"."+bucket, strconv.FormatUint(val, 10))
}
