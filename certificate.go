package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type CertResult struct {
	Domain     string
	CNAME      string
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
	var (
		ret = CertResult{
			Domain: domain,
			Status: "Error",
		}
		et    time.Time
		err   error
		dom   string
		cname string
	)
	dparts := strings.Split(domain, "|")
	if len(dparts) == 2 {
		ret.Domain = dparts[0]
		ret.CNAME = dparts[1]
		dom = dparts[0]
		cname = dparts[1]
	} else {
		ret.Domain = domain
		ret.CNAME = domain
		dom = domain
		cname = domain
	}

	for i := 1; i < 4; i++ {
		et, err = dc.GetExpireTime(dom, cname)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}

	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%s", err)
		return ret
	}

	days := int(et.Sub(time.Now()).Hours() / 24)
	log.Println("[INFO] Certificate", dom, "Expire After", days, "Days,", et)
	ret.Status = "OK"
	ret.ExpireAt = et
	ret.ExpireDays = days
	return ret
}

func (dc *CertificatesChecker) GetExpireTime(domain string, cname string) (time.Time, error) {
	var dialer net.Dialer
	dialer.Timeout = 5 * time.Second
	cfg := &tls.Config{
		ServerName: domain,
	}
	conn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:443", cname), cfg)
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
