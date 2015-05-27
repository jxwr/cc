package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jxwr/cc/frontend/api"
)

func do(method, url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	reqJson, _ := json.Marshal(in)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	client.Timeout = timeout
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		var resp api.Response
		err = json.Unmarshal(body, &resp)
		return &resp, err
	} else {
		return nil, err
	}
}

func HttpPost(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("POST", url, in, timeout)
}

func HttpPut(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("PUT", url, in, timeout)
}

func HttpGet(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("GET", url, in, timeout)
}
