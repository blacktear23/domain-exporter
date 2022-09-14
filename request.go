package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type RequestResult struct {
	Domain     string
	Status     string
	Host       string
	Path       string
	Address    string
	StatusCode int
	ErrorMsg   string
}

type RequestParams struct {
	Domain string
	Host   string
	Path   string
	Https  bool
}

type RequestResults map[string]RequestResult

type RequestChecker struct {
	Domains []RequestConfig
}

func NewRequestChecker(domains []RequestConfig) *RequestChecker {
	return &RequestChecker{
		Domains: domains,
	}
}

func (rc *RequestChecker) Check() RequestResults {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
		ret  RequestResults = make(RequestResults)
	)
	domains := []*RequestParams{}
	for _, cfg := range rc.Domains {
		for _, domain := range cfg.Domains {
			domains = append(domains, &RequestParams{
				Domain: domain,
				Host:   cfg.Host,
				Path:   cfg.Path,
				Https:  cfg.Https,
			})
		}
	}
	wg.Add(len(domains))
	for _, item := range domains {
		go func(params *RequestParams) {
			rr := rc.CheckOneDomain(params)
			lock.Lock()
			key := fmt.Sprintf("%s @ %s", params.Host, params.Domain)
			ret[key] = rr
			lock.Unlock()
			if rr.ErrorMsg != "" {
				log.Printf("RequestChecker Error: %s: %s%s -> %v", params.Domain, params.Host, params.Path, rr.ErrorMsg)
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ret
}

func (rc *RequestChecker) CheckOneDomain(params *RequestParams) RequestResult {
	ret := RequestResult{
		Domain:     params.Domain,
		Status:     "Error",
		Host:       params.Host,
		Path:       params.Path,
		Address:    "",
		StatusCode: 0,
	}
	addrs, err := net.LookupHost(params.Domain)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%v", err)
		return ret
	}
	if len(addrs) == 0 {
		ret.ErrorMsg = "Domain has no IP addresses"
		return ret
	}
	addr := selectAddress(addrs)
	ret.Address = addr
	responseOk, statusCode, err := rc.doRequest(addr, params)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%v", err)
	}
	if responseOk {
		ret.Status = "OK"
	}
	ret.StatusCode = statusCode
	return ret
}

func selectAddress(addrs []string) string {
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			// Return first IPv4 address
			return addr
		}
	}
	// No IPv4 Address just return first
	return addrs[0]
}

func (rc *RequestChecker) doRequest(raddr string, params *RequestParams) (bool, int, error) {
	var url string
	if params.Https {
		url = fmt.Sprintf("https://%s%s", params.Host, params.Path)
	} else {
		url = fmt.Sprintf("http://%s%s", params.Host, params.Path)
	}
	// Generate request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}
	req.Host = params.Host
	req.Header.Add("Host", params.Host)

	// Prepare for http client
	tp := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := net.Dialer{}
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			return dialer.DialContext(ctx, network, fmt.Sprintf("%s:%s", raddr, port))
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tp,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		buf := make([]byte, 1)
		_, err := resp.Body.Read(buf)
		return true, 200, err
	}
	return false, resp.StatusCode, fmt.Errorf("Status not equals to 200, %v", resp.StatusCode)
}

func (rc *RequestChecker) RequestHttp(addr string, params *RequestParams) (bool, int, error) {
	var url string
	if params.Https {
		url = fmt.Sprintf("https://%s%s", addr, params.Path)
	} else {
		url = fmt.Sprintf("http://%s%s", addr, params.Path)
	}
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}
	req.Host = params.Host
	req.Header.Add("Host", params.Host)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		buf := make([]byte, 1)
		_, err := resp.Body.Read(buf)
		return true, 200, err
	}
	return false, resp.StatusCode, fmt.Errorf("Status not equals to 200, %v", resp.StatusCode)
}
