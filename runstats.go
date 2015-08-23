package runstats

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/bmhatfield/go-runtime-metrics/collector"
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

	config := collector.DefaultConfig
	config.Addr = *statsd
	config.Prefix = *prefix
	config.PauseDur = time.Duration(*pause) * time.Second
	config.EnableCPU = *cpu
	config.EnableMem = *mem
	config.EnableGC = *gc

	c, err := collector.New(config)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to Statsd on %s - %s", *statsd, err))
	}
	c.Run()
}
