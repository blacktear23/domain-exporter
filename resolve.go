package main

import (
	"fmt"
	"net"
	"sync"
)

type ResolveResult struct {
	Domain   string
	Status   string
	IPs      []string
	ErrorMsg string
}

type ResolveResults map[string]ResolveResult

type ResolveChecker struct {
	Domains []string
}

func NewResolveChecker(domains []string) *ResolveChecker {
	return &ResolveChecker{
		Domains: domains,
	}
}

func (rc *ResolveChecker) Check() ResolveResults {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
		ret  ResolveResults = make(ResolveResults)
	)
	wg.Add(len(rc.Domains))
	for _, item := range rc.Domains {
		go func(domain string) {
			rr := rc.CheckOneDomain(domain)
			lock.Lock()
			ret[domain] = rr
			lock.Unlock()
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ret
}

func (rc *ResolveChecker) CheckOneDomain(domain string) ResolveResult {
	ret := ResolveResult{
		Domain:   domain,
		Status:   "Error",
		IPs:      []string{},
		ErrorMsg: "",
	}

	addrs, err := net.LookupHost(domain)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%v", err)
	}
	ret.IPs = addrs
	if len(addrs) > 0 {
		ret.Status = "OK"
	}
	return ret
}
