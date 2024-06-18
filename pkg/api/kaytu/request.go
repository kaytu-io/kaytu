package kaytu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

var ErrLogin = errors.New("your session is expired, please login")

func ApiKeyRequest(accessToken string) (*CreateAPIKeyResponse, error) {
	id, _ := uuid.NewV7()
	request := CreateAPIKeyRequest{Name: id.String()}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.kaytu.io/kaytu/auth/api/v1/key/create", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("[ConfigurationRequest]: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ConfigurationRequest]: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[ConfigurationRequest]: %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[ConfigurationRequest]: %v", err)
	}

	if res.StatusCode == 401 {
		return nil, ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [ConfigurationRequest]: %s", res.StatusCode, string(body))
	}

	response := CreateAPIKeyResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[ConfigurationRequest]: %v", err)
	}
	return &response, nil
}
