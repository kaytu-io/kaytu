package server

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"strings"
	"time"
)

type Config struct {
	AccessToken      string `json:"access-token"`
	DefaultWorkspace string `json:"default_workspace"`
}

var ExpiredSession = fmt.Errorf("your session has expired, please login again using `kaytu login`")

func GetConfig() (*Config, error) {
	home := os.Getenv("HOME")
	data, err := os.ReadFile(home + "/.kaytu/kaytu-config.json")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, nil
		}
		return nil, fmt.Errorf("[CredentialsFile] : %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("[getConfig] : %v", err)
	}

	if config.AccessToken == "" {
		return nil, fmt.Errorf("please log in first")
	}

	checkEXP, err := CheckExpirationTime(config.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("[getConfig] : %v", err)
	}

	if checkEXP == true {
		return nil, ExpiredSession
	}
	return &config, nil
}

func RemoveConfig() error {
	home := os.Getenv("HOME")
	err := os.Remove(home + "/.kaytu/kaytu-config.json")
	if err != nil {
		return fmt.Errorf("[removeConfig] : %v", err)
	}
	return nil
}

func SetConfig(data Config) error {
	configs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("[addConfig] : %v", err)
	}
	home := os.Getenv("HOME")
	_, err = os.Stat(home + "/.kaytu")
	if err != nil {
		err = os.Mkdir(home+"/.kaytu", os.ModePerm)
		if err != nil {
			return fmt.Errorf("[addConfig] : %v", err)
		}
	}

	err = os.WriteFile(home+"/.kaytu/kaytu-config.json", configs, os.ModePerm)
	if err != nil {
		return fmt.Errorf("[addConfig] : %v", err)
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
