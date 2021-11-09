package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/HotPotatoC/kvstore-rewrite/logger"
	"github.com/spf13/viper"
)

const (
	defaultENVPrefix      = "KVSTORE"
	defaultConfigFileName = "kvstore.toml"
)

// Defaults is the default configuration values and
// is used when a configuration file was not found
var Defaults = map[string]interface{}{
	"server.port":  7275,
	"server.addrs": []string{"tcp://127.0.0.1"},

	"database.path": "./dump.kvsdb",

	"log.level":       0,
	"log.file_path":   "/var/log/kvstore/kvstore-server.log",
	"log.max_size":    50 * 1024 * 1024,
	"log.max_backups": 10,
	"log.max_age":     30,
	"log.compress":    true,
}

func setDefaults() {
	for key, defaultValue := range Defaults {
		if viper.Get(key) == nil {
			viper.SetDefault(key, defaultValue)
		}
	}
}

// Load loads the configuration from the given path
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
