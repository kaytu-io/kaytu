package kaytu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrLogin = errors.New("your session is expired, please login")

func CreateApiKeyRequest(accessToken, name string) (*CreateAPIKeyResponse, error) {
	request := CreateAPIKeyRequest{Name: name}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.kaytu.io/kaytu/auth/api/v1/key/create", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("[CreateApiKeyRequest]: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[CreateApiKeyRequest]: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[CreateApiKeyRequest]: %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[CreateApiKeyRequest]: %v", err)
	}

	if res.StatusCode == 401 {
		return nil, ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [CreateApiKeyRequest]: %s", res.StatusCode, string(body))
	}

	response := CreateAPIKeyResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[CreateApiKeyRequest]: %v", err)
	}
	return &response, nil
}

func ListApiKeyRequest(accessToken string) ([]ApiKey, error) {
	req, err := http.NewRequest("GET", "https://api.kaytu.io/auth/api/v1/keys", nil)
	if err != nil {
		return nil, fmt.Errorf("[ListApiKeyRequest]: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ListApiKeyRequest]: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[ListApiKeyRequest]: %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[ListApiKeyRequest]: %v", err)
	}

	if res.StatusCode == 401 {
		return nil, ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [ListApiKeyRequest]: %s", res.StatusCode, string(body))
	}

	var response []ApiKey
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[ListApiKeyRequest]: %v", err)
	}
	return response, nil
}

func DeleteApiKeyRequest(accessToken, name string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://api.kaytu.io/auth/api/v1/key/%s/delete", name), nil)
	if err != nil {
		return fmt.Errorf("[DeleteApiKeyRequest]: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[DeleteApiKeyRequest]: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("[DeleteApiKeyRequest]: %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf("[DeleteApiKeyRequest]: %v", err)
	}

	if res.StatusCode == 401 {
		return ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return fmt.Errorf("server returned status code %d, [DeleteApiKeyRequest]: %s", res.StatusCode, string(body))
	}

	return nil
}
