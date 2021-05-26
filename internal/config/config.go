package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HotPotatoC/kvstore/internal/logger"
	"github.com/spf13/viper"
)

const (
	defaultENVPrefix      = "KVSTORE"
	defaultConfigFileName = "kvstore.toml"
)

// Defaults is the default configuration values and
// is used when a configuration file was not found
var Defaults = map[string]interface{}{
	"server.host":                    "0.0.0.0",
	"server.port":                    7275,
	"server.protocol":                "tcp",
	"server.multicore":               false,
	"server.reuse_port":              false,
	"server.read_buffer_cap":         0x2000000, // 32mb
	"server.tcp_keep_alive":          true,
	"server.tcp_keep_alive_duration": 10 * time.Minute,

	"log.path":        "/var/log/kvstore/kvstore-server.log",
	"log.level":       -1,
	"log.max_size":    10,
	"log.max_backups": 6,
	"log.max_age":     28,
	"log.compress":    true,

	"aof.enabled":       true,
	"aof.path":          "./kvstore-aof.log",
	"aof.persist_after": time.Minute,
}

func setDefaults() {
	for key, defaultValue := range Defaults {
		if viper.Get(key) == nil {
			viper.SetDefault(key, defaultValue)
		}
	}
}

func Load(path ...string) error {
	var pathToFile string

	switch {
	case len(path) < 1, path[0] == "":
		// Use the current working directory if no path was provided
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		pathToFile = filepath.Join(wd, defaultConfigFileName)
	default:
		p, err := filepath.Abs(filepath.Clean(path[0]))
		if err != nil {
			return err
		}

		pathToFile = p
	}

	base := filepath.Base(pathToFile)
	dir := filepath.Dir(pathToFile)
	cfgFile := strings.Split(base, ".")

	viper.SetConfigName(cfgFile[0])
	viper.SetConfigType(cfgFile[1])
	viper.AddConfigPath(dir)

	viper.AutomaticEnv()

	logger.S().Debug("loading config file...")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.S().Debug("config file not found")
			logger.S().Debug("using default configs")
			setDefaults()

			return nil
		}

		return err
	}

	logger.S().Debugf("config file found: %s", pathToFile)
	logger.S().Debug("setting default values for unconfigured config keys")
	setDefaults()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix(defaultENVPrefix)

	return nil
}
