package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ksarch-saas/cc/frontend/api"
)

type ExtraHeader struct {
	User  string
	Role  string
	Token string
}

func do(method, url string, in interface{}, timeout time.Duration, extra *ExtraHeader) (*api.Response, error) {
	reqJson, _ := json.Marshal(in)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if extra != nil {
		if extra.User != "" {
			req.Header.Set("User", extra.User)
		}
		if extra.Role != "" {
			req.Header.Set("Role", extra.Role)
		}
		if extra.Token != "" {
			req.Header.Set("Token", extra.Token)
		}
	}

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
		var rsp api.Response
		d := json.NewDecoder(bytes.NewReader(body))
		d.UseNumber()
		err = d.Decode(&rsp)
		return &rsp, err
	} else {
		return nil, err
	}
}

func HttpPost(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("POST", url, in, timeout, nil)
}

func HttpPut(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("PUT", url, in, timeout, nil)
}

func HttpGet(url string, in interface{}, timeout time.Duration) (*api.Response, error) {
	return do("GET", url, in, timeout, nil)
}

func HttpPostExtra(url string, in interface{}, timeout time.Duration, extra *ExtraHeader) (*api.Response, error) {
	return do("POST", url, in, timeout, extra)
}

func HttpPutExtra(url string, in interface{}, timeout time.Duration, extra *ExtraHeader) (*api.Response, error) {
	return do("PUT", url, in, timeout, extra)
}

func HttpGetExtra(url string, in interface{}, timeout time.Duration, extra *ExtraHeader) (*api.Response, error) {
	return do("GET", url, in, timeout, extra)
}
