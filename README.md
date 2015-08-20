# go-runtime-metrics
Collect Golang Runtime Metrics, outputting to a stats handler (currently, statsd)

The intent of this library is to be a "side effect" import. You can kick off the collector merely by importing this into your main:

`import _ "github.com/bmhatfield/go-runtime-metrics"`

This library has a few optional flags it depends on. It won't be able to output stats until you call `flag.Parse()`, which is generally done in your `func main() {}`.

Once imported and running, you can expect a number of Go runtime metrics to be sent to statsd over UDP. An example of what this looks like:

![Dashboard Screenshot](/screenshot.png?raw=true)
