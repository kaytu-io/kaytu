package wastage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/server"
	"io"
	"net/http"
)

var ErrLogin = errors.New("your session is expired, please login")

func Ec2InstanceWastageRequest(reqBody EC2InstanceWastageRequest) (*EC2InstanceWastageResponse, error) {
	config, err := server.GetConfig()
	if err != nil {
		return nil, err
	}

	payloadEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.kaytu.io/kaytu/wastage/api/v1/wastage/ec2-instance", bytes.NewBuffer(payloadEncoded))
	//req, err := http.NewRequest("POST", "http://localhost:8000/api/v1/wastage/ec2-instance", bytes.NewBuffer(payloadEncoded))
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	req.Header.Add("content-type", "application/json")
	if config != nil && len(config.AccessToken) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.AccessToken))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}

	if res.StatusCode == 403 {
		return nil, ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [requestAbout] : %s", res.StatusCode, string(body))
	}

	response := EC2InstanceWastageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	return &response, nil
}
