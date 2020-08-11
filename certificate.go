package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type CertResult struct {
	Domain     string
	Status     string
	ErrorMsg   string
	ExpireAt   time.Time
	ExpireDays int
}

type CertResults map[string]CertResult

type CertificatesChecker struct {
	Domains []string
}

func NewCertificatesChecker(domains []string) *CertificatesChecker {
	return &CertificatesChecker{
		Domains: domains,
	}
}

func (dc *CertificatesChecker) Check() CertResults {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
		ret  CertResults = make(CertResults)
	)
	wg.Add(len(dc.Domains))
	for _, item := range dc.Domains {
		go func(domain string) {
			cr := dc.CheckOneDomain(domain)
			lock.Lock()
			ret[domain] = cr
			lock.Unlock()
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ret
}

func (dc *CertificatesChecker) CheckOneDomain(domain string) CertResult {
	ret := CertResult{
		Domain: domain,
		Status: "Error",
	}
	et, err := dc.GetExpireTime(domain)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%s", err)
		return ret
	}
	days := int(et.Sub(time.Now()).Hours() / 24)
	log.Println("[INFO] Certificate", domain, "Expire After", days, "Days,", et)
	ret.Status = "OK"
	ret.ExpireAt = et
	ret.ExpireDays = days
	return ret
}

func (dc *CertificatesChecker) GetExpireTime(domain string) (time.Time, error) {
	var dialer net.Dialer
	dialer.Timeout = 5 * time.Second
	conn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:443", domain), nil)
	if err != nil {
		return time.Time{}, err
	}
	defer conn.Close()
	connStat := conn.ConnectionState()
	for _, cert := range connStat.PeerCertificates {
		if !cert.IsCA {
			return cert.NotAfter, nil
		}
	}
	return time.Time{}, fmt.Errorf("Invalid certificate: no peer certificates")
}
