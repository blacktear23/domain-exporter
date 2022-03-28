package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type DNSProber struct {
	config *Config
}

type DNSProbeResult struct {
	Domain   string   `json:'domain'`
	IPs      []string `json:'ips'`
	ErrorMsg string   `json:'error_msg'`
}

func NewDNSProber(cfg *Config) *DNSProber {
	return &DNSProber{
		config: cfg,
	}
}

func (p *DNSProber) Probe() ([]byte, error) {
	domains := p.config.ResolveDomains
	result := p.probeDomains(domains)
	return json.MarshalIndent(result, "", "\t")
}

func (p *DNSProber) probeDomains(domains []string) []DNSProbeResult {
	ret := make([]DNSProbeResult, len(domains))
	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(domains))
	for i, domain := range domains {
		go func(idx int, domainName string) {
			result := p.probeDomain(domainName)
			lock.Lock()
			ret[idx] = result
			lock.Unlock()
			wg.Done()
		}(i, domain)
	}
	wg.Wait()
	return ret
}

func (p *DNSProber) probeDomain(domain string) DNSProbeResult {
	ret := DNSProbeResult{
		Domain:   domain,
		IPs:      []string{},
		ErrorMsg: "",
	}

	addrs, err := net.LookupHost(domain)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%v", err)
	}
	ret.IPs = addrs
	return ret
}
