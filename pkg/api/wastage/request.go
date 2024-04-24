package wastage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Ec2InstanceWastageRequest(reqBody EC2InstanceWastageRequest) (*EC2InstanceWastageResponse, error) {
	payloadEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.kaytu.io/kaytu/wastage/api/v1/wastage/ec2-instance", bytes.NewBuffer(payloadEncoded))
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	req.Header.Add("content-type", "application/json")
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
	response := EC2InstanceWastageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	return &response, nil
}
