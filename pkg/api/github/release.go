package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetLatestRelease(repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("[release] : %v", err)
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[release] : %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[release] : %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[release] : %v", err)
	}

	if res.StatusCode == 404 {
		return nil, fmt.Errorf("repository %s not found", repo)
	}
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [release] : %s", res.StatusCode, string(body))
	}

	response := Release{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[release] : %v", err)
	}
	return &response, nil

}
