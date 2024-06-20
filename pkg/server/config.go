package server

import (
	"encoding/json"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/version"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	AccessToken     string    `json:"access-token"`
	Plugins         []*Plugin `json:"plugins"`
	LastUpdateCheck time.Time `json:"lastUpdateCheck"`
	LastVersion     string    `json:"lastVersion"`
}

var (
	once      sync.Once
	config    *Config
	configErr error
)

func GetConfig() (*Config, error) {
	once.Do(func() {
		config, configErr = loadConfig()
	})
	return config, configErr
}

func loadConfig() (*Config, error) {
	config := Config{
		AccessToken:     "",
		Plugins:         nil,
		LastUpdateCheck: time.Time{},
		LastVersion:     version.VERSION,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("[loadConfig]: %v", err)
	}

	path := filepath.Join(home, ".kaytu", "kaytu-config.json")

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// if the file does not exist, return nil
			return &config, nil
		}
		return nil, fmt.Errorf("[loadConfig]: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[loadConfig]: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("[loadConfig]: %v", err)
	}

	if config.AccessToken != "" {
		checkEXP, err := CheckExpirationTime(config.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("[loadConfig]: %v", err)
		}

		if checkEXP == true {
			config.AccessToken = ""
		}
	}
	return &config, nil
}

func RemoveConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("[removeConfig]: %v", err)
	}
	err = os.Remove(filepath.Join(home, ".kaytu", "kaytu-config.json"))
	if err != nil {
		return fmt.Errorf("[removeConfig]: %v", err)
	}
	return nil
}

func SetConfig(data Config) error {
	configs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("[SetConfig]: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("[SetConfig]: %v", err)
	}
	_, err = os.Stat(filepath.Join(home, ".kaytu"))
	if err != nil {
		err = os.Mkdir(filepath.Join(home, ".kaytu"), os.ModePerm)
		if err != nil {
			return fmt.Errorf("[SetConfig]: %v", err)
		}
	}

	err = os.WriteFile(filepath.Join(home, ".kaytu", "kaytu-config.json"), configs, os.ModePerm)
	if err != nil {
		return fmt.Errorf("[SetConfig]: %v", err)
	}
	return nil
}

func CheckExpirationTime(accessToken string) (bool, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, err
	}

	if _, ok := claims["exp"]; !ok {
		return false, nil
	}

	var tm time.Time
	switch iat := claims["exp"].(type) {
	case float64:
		tm = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tm = time.Unix(v, 0)
	}
	timeNow := time.Now()
	if tm.Before(timeNow) {
		return true, nil
	} else if tm.After(timeNow) {
		return false, nil
	} else {
		return true, err
	}
}
