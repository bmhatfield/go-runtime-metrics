package runstats

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/bmhatfield/go-runtime-metrics/collector"
	"github.com/peterbourgon/g2s"
)

var (
	statsd *string = flag.String("statsd", "localhost:8125", "Statsd host:port pair")
	prefix *string = flag.String("metric-prefix", "<detect-hostname>", "Metric prefix path; detects the local hostname by default")
	pause  *int    = flag.Int("pause", 10, "Collection pause interval")

	publish *bool = flag.Bool("publish-runtime-stats", true, "Collect go runtime statistics")
	cpu     *bool = flag.Bool("cpu", true, "Collect CPU Statistics")
	mem     *bool = flag.Bool("mem", true, "Collect Memory Statistics")
	gc      *bool = flag.Bool("gc", true, "Collect GC Statistics (requires Memory be enabled)")

	s g2s.Statter
)

func init() {
	go runCollector()
}

func runCollector() {
	for !flag.Parsed() {
		// Defer execution of this goroutine.
		runtime.Gosched()

		// Add an initial delay while the program initializes to avoid attempting to collect
		// metrics prior to our flags being available / parsed.
		time.Sleep(1 * time.Second)
	}

	s, err := g2s.Dial("udp", *statsd)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to Statsd on %s - %s", *statsd, err))
	}

	if *prefix == "<detect-hostname>" {
		hn, err := os.Hostname()

		if err != nil {
			*prefix = "go.unknown"
		} else {
			*prefix = "go." + hn
		}
	}
	*prefix += "."

	gaugeFunc := func(key string, val uint64) {
		s.Gauge(1.0, *prefix+key, strconv.FormatUint(val, 10))
	}
	c := collector.New(gaugeFunc)
	c.PauseDur = time.Duration(*pause) * time.Second
	c.EnableCPU = *cpu
	c.EnableMem = *mem
	c.EnableGC = *gc

	if *publish {
		c.Run()
	}
}
