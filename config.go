package drone

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
)

func GetConfig(path string) *Config {
	conf := &Config{}
	good := false

	// get outer config dir
	d := filepath.Dir(path)

	// check if main config exists
	_, err := os.Stat(path)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			log.Print("config: base config not found")
		} else {
			log.Panicf("config: stat err :: %v", err)
		}
	} else {
		_, err = toml.DecodeFile(path, conf)
		if err != nil {
			log.Panicf("config: decode err :: %v", err)
		} else {
			log.Printf("config: loaded root: %s", path)
			good = true
		}
	}

	// check conf.d
	cd, err := os.Stat(d)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			log.Printf("config: %s doesn't exist, skipping")
		} else {
			log.Panicf("config: stat err :: %v", err)
		}
	} else {
		if cd.IsDir() {
			cdf, _ := filepath.Glob(filepath.Join(d, "conf.d", "*.toml"))

			for _, c := range cdf {
				_, err = toml.DecodeFile(c, conf)
				if err != nil {
					log.Printf("config: decode err (file: %s) :: %v", c, err)
				} else {
					log.Printf("config: loaded conf.d: %s", c)
					good = true
				}
			}
		} else {
			log.Printf("config: %s isn't a directory, skipping")
		}
	}

	if !good {
		log.Fatal("config: i didn't load any sort of config. exiting.")
	}

	return conf
}

type Config struct {
	Graylog   graylogConfig
	Collector collectorConfig
	Logs      logsConfig
}

type graylogConfig struct {
	Address string
}

type collectorConfig struct {
	Hostname string
}

type logsConfig map[string]LogConfig

type LogConfig struct {
	File      string
	Parser    string
	Pattern   string
	ShortText string `toml:"short_text",omitempty`
}
