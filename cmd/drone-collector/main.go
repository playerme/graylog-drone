package main

import (
	"github.com/hpcloud/tail"
	flags "github.com/jessevdk/go-flags"
	"github.com/player-me/graylog-drone"
	"log"
	"runtime"
	"sync"
)

var verbose bool = false

func main() {
	log.Println("starting drone")

	// set workers on max. this is just simple.
	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	// flags definition
	var opts struct {
		ConfigPath string `short:"c" long:"config-path" description:"Path to config file" default:"/etc/drone/drone.toml"`
		Verbose    bool   `short:"v" description:"be louder"`
	}

	// parse flags
	flags.Parse(&opts)

	verbose = opts.Verbose

	// grab the config
	config := drone.GetConfig(opts.ConfigPath)

	if verbose {
		log.Printf("will be sending logs to %s", config.Graylog.Address)
	}

	// create the graylog connection
	collector, err := drone.NewCollector(config)
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}

	// a sync helper to keep the goroutines open
	var wg sync.WaitGroup

	// add the number of goroutines
	wg.Add(len(config.Logs))

	// spawn the goroutines
	for logName, logConfig := range config.Logs {
		go tailer(collector, logName, logConfig, wg)
	}

	// wait for all goroutines to finish before exiting
	wg.Wait()

	log.Println("stopped drone")
}

func tailer(collector *drone.Collector, name string, logConfig drone.LogConfig, wg sync.WaitGroup) {

	log.Printf("started %s with file %s\n", name, logConfig.File)

	// when this function exits, tell sync we're done
	defer wg.Done()

	// start tailing file
	t, err := tail.TailFile(logConfig.File, tail.Config{
		Follow: true, // -f - keeps reading infinitely
		ReOpen: true, // -F - keeps the file open when it actually moves
		Location: &tail.SeekInfo{
			Whence: 2, // skips re-sending the most recent line on startup
		},
	})
	if err != nil {
		log.Printf("[%s]: tail open on file %s failed: %v", name, logConfig.File, err)
	}

	// also when this function exits, cleanup the inotify hooks
	defer t.Cleanup()

	// build some extra data we can query in graylog
	metadata := map[string]interface{}{
		"log_name": name,
		"log_file": logConfig.File,
	}

	// on every line emitted, write to graylog.
	for line := range t.Lines {
		text := line.Text
		data := make(map[string]interface{})

		if verbose {
			log.Printf("[%s] %s\n", name, text)
		}

		if logConfig.Parser != "none" {
			text, data, err = parseLog(text, logConfig)
			if err != nil {
				log.Printf("[%s] parser error: %v", err)
			}
		}

		// mix in the outer metadata map
		for k, v := range metadata {
			data[k] = v
		}

		// include original raw log
		data["raw_line"] = line.Text

		// write the line!
		collector.Write(text, data)
	}

}

func parseLog(text string, config drone.LogConfig) (out string, data map[string]interface{}, err error) {

	switch config.Parser {
	case "grok":
		out, data, err = drone.GrokParser(text, config)
		break
	case "none":
		return text, map[string]interface{}{}, nil
	}

	return out, data, err

}
