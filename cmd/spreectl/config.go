package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ralfonso/spree/auth"
	"github.com/uber-go/zap"
)

// getConfig gets the client config from disk.
func getConfig(ll zap.Logger) (*auth.ClientConfig, error) {
	configDir := configHome()
	configFile := configFileName(configDir)
	f, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	conf := &auth.ClientConfig{}
	err = json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// storeConfig stores the client config to disk.
func storeConfig(conf *auth.ClientConfig, ll zap.Logger) error {
	jsonConf, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	configDir := configHome()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := configFileName(configDir)
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(jsonConf)
	return err
}

func configHome() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir(), ".config", "spreectl")
	}

	return configHome
}

func configFileName(configDir string) string {
	return filepath.Join(configDir, defaultAccessTokenFileName)
}
