package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var JsonHead = map[string]string{
	"Content-Type": "application/json",
}

type ProxyConfig struct {
	ID       int64  `json:"id"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Country  string `json:"country"`
	Username string `json:"username"`
	Pass     string `json:"pass"`
}

func ProxyGetRequestAll(requestUrl string, heads map[string]string, config *ProxyConfig) (responseData []byte,
	cookie map[string]string, err error) {
	var client http.Client
	if config != nil && config.Ip != "" {
		proxyStr := ""
		if config.Username == "" {
			proxyStr = fmt.Sprintf("%s://%s:%d", config.Protocol, config.Ip, config.Port)
		} else {
			proxyStr = fmt.Sprintf("%s://%s:%s@%s:%d", config.Protocol, config.Username, config.Pass, config.Ip, config.Port)
		}
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			return nil, cookie, err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client = http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		}
	} else {
		client = http.Client{}
	}
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, cookie, err
	}
	if heads != nil {
		for key, value := range heads {
			req.Header.Set(key, value)
		}
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, cookie, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, cookie, err
	}
	cookie = make(map[string]string)
	for key, values := range response.Header {
		if len(values) == 1 {
			cookie[key] = values[0]
		} else if len(values) > 1 {
			data := ""
			for _, value := range values {
				data = fmt.Sprintf("%s;%s=%s", data, key, value)
			}
			cookie[key] = data
		} else {
			cookie[key] = ""
		}
	}
	return body, cookie, nil
}

func ProxyPostRequest(requestUrl string, heads map[string]string, requestData []byte,
	config *ProxyConfig) (responseData []byte, cookie map[string]string, err error) {
	var client http.Client
	if config != nil && config.Ip != "" {
		proxyStr := ""
		if config.Username == "" {
			proxyStr = fmt.Sprintf("%s://%s:%d", config.Protocol, config.Ip, config.Port)
		} else {
			proxyStr = fmt.Sprintf("%s://%s:%s@%s:%d", config.Protocol, config.Username, config.Pass, config.Ip, config.Port)
		}
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			return nil, cookie, err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client = http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		}
	} else {
		client = http.Client{}
	}
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, cookie, err
	}
	if heads != nil {
		for key, value := range heads {
			req.Header.Set(key, value)
		}
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, cookie, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, cookie, err
	}
	cookie = make(map[string]string)
	for key, values := range response.Header {
		if len(values) == 1 {
			cookie[key] = values[0]
		} else if len(values) > 1 {
			data := ""
			for _, value := range values {
				data = fmt.Sprintf("%s;%s=%s", data, key, value)
			}
			cookie[key] = data
		} else {
			cookie[key] = ""
		}
	}
	return body, cookie, nil
}
