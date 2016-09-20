package drone

import (
	"github.com/Graylog2/go-gelf/gelf"
	"time"
)

// A Collector is a metastructure to talk to Graylog via GELF.
type Collector struct {
	writer *gelf.Writer
	host   string
}

// New creates a Collector instance to talk to Graylog via GELF.
// It takes the global config object, pulls the node and gelf sections,
// and sets those in the Collector.
func NewCollector(config *Config) (coll *Collector, err error) {
	var gelfWriter *gelf.Writer

	if gelfWriter, err = gelf.NewWriter(config.Graylog.Address); err != nil {
		return nil, err
	}

	coll = new(Collector)
	coll.writer = gelfWriter
	coll.host = config.Collector.Hostname

	return coll, nil
}

// Write creates a gelf.Message with forced data.
func (coll *Collector) Write(msg string, extra map[string]interface{}) (err error) {

	m := gelf.Message{
		Version:  "1.1",
		Host:     coll.host,
		Short:    msg,
		TimeUnix: float64(time.Now().Unix()),
		Level:    6, // info always
		Facility: "drone",
		Extra:    extra,
	}

	if err = coll.writer.WriteMessage(&m); err != nil {
		return err
	}

	return nil

}
