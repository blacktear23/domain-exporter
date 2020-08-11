package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func GetWhoisTimeout(domain string, timeout time.Duration) (result string, err error) {

	var (
		parts      []string
		zone       string
		buffer     []byte
		connection net.Conn
	)
	parts = strings.Split(domain, ".")
	if len(parts) < 2 {
		err = fmt.Errorf("Domain(%s) name is wrong!", domain)
		return
	}
	//last part of domain is zome
	zone = parts[len(parts)-1]
	server, ok := servers[zone]
	if !ok {
		err = fmt.Errorf("No such server for zone %s. Domain %s.", zone, domain)
		return
	}
	connection, err = net.DialTimeout("tcp", net.JoinHostPort(server, "43"), timeout)
	if err != nil {
		//return net.Conn error
		return
	}
	defer connection.Close()

	connection.Write([]byte(domain + "\r\n"))
	buffer, err = ioutil.ReadAll(connection)
	if err != nil {
		return
	}
	result = string(buffer[:])
	return
}

type WhoisResult struct {
	Domain     string
	Status     string
	ErrorMsg   string
	ExpireAt   time.Time
	ExpireDays int
}

type WhoisResults map[string]WhoisResult

type WhoisChecker struct {
	Domains []string
}

func NewWhoisChecker(domains []string) *WhoisChecker {
	return &WhoisChecker{
		Domains: domains,
	}
}

func (wc *WhoisChecker) Check() WhoisResults {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
		ret  WhoisResults = make(WhoisResults)
	)
	wg.Add(len(wc.Domains))
	for _, item := range wc.Domains {
		go func(domain string) {
			cr := wc.CheckOneDomain(domain)
			lock.Lock()
			ret[domain] = cr
			lock.Unlock()
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ret
}

func (wc *WhoisChecker) CheckOneDomain(domain string) WhoisResult {
	ret := WhoisResult{
		Domain: domain,
		Status: "Error",
	}
	whois, err := GetWhoisTimeout(domain, 5*time.Second)
	if err != nil {
		ret.ErrorMsg = fmt.Sprintf("%s", err)
		return ret
	}
	et, days := wc.decodeWhoisInfo(whois)
	log.Println("[INFO] Whois", domain, "Expire After", days, "Days,", et)
	ret.Status = "OK"
	ret.ExpireAt = et
	ret.ExpireDays = days
	return ret
}

func (wc *WhoisChecker) decodeWhoisInfo(info string) (time.Time, int) {
	for _, rline := range strings.Split(info, "\n") {
		if strings.Contains(rline, "Expir") {
			line := strings.TrimSpace(rline)
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				expireDate := strings.TrimSpace(parts[1])
				expireDate = strings.ReplaceAll(expireDate, "T", " ")
				eparts := strings.Split(strings.ToUpper(expireDate), "Z")
				et := parseDateStr(eparts[0])
				days := int(et.Sub(time.Now()).Hours() / 24)
				return et, days
			}
		}
	}

	return time.Time{}, 0
}

func parseDateStr(date string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		log.Println("[Error]", err)
		return time.Time{}
	}
	return t
}
