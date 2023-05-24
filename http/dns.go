package http

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var dnsResolvers = []string{
	"1.1.1.1", // Cloudflare
	"1.0.0.1",
	"8.8.8.8", // Google
	"8.8.4.4",
	"9.9.9.9", // Quad9
	"149.112.112.112",
	"208.67.222.222", // OpenDNS Cisco
	"208.67.220.220",
	"208.67.222.220",
	"208.67.220.222",
}
var currentResolverIndex uint64

type DNSRecord struct {
	IP          string  `json:"ip"`
	Name        string  `json:"name"`
	ASNumber    int     `json:"as_number"`
	ASOrg       string  `json:"as_org"`
	CountryID   string  `json:"country_id"`
	City        string  `json:"city"`
	Version     string  `json:"version"`
	Error       string  `json:"error"`
	DNSSEC      bool    `json:"dnssec"`
	Reliability float64 `json:"reliability"`
	CheckedAt   string  `json:"checked_at"`
	CreatedAt   string  `json:"created_at"`
}

func FetchReliableDNSRecords(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("failed to fetch data: " + err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body: " + err.Error())
	}

	var records []DNSRecord
	err = json.Unmarshal(body, &records)
	if err != nil {
		return nil, errors.New("failed to unmarshal JSON: " + err.Error())
	}

	var reliableIPs []string
	var wg sync.WaitGroup
	ipChan := make(chan string)

	for _, record := range records {
		if record.Reliability == 1 {
			ip := net.ParseIP(record.IP)
			if ip != nil && ip.To4() != nil {
				wg.Add(1)
				go func(ip string) {
					defer wg.Done()
					if checkDNS(ip) {
						ipChan <- ip
					}
				}(record.IP)
			}
		}
	}

	go func() {
		wg.Wait()
		close(ipChan)
	}()

	for ip := range ipChan {
		reliableIPs = append(reliableIPs, ip)
	}

	log.Printf("Reliable DNS servers found: %d", len(reliableIPs))

	return reliableIPs, nil
}

func checkDNS(ip string) bool {
	r := &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * 5000,
			}
			return d.DialContext(ctx, "udp", ip+":53")
		},
	}

	_, err := r.LookupHost(context.Background(), "www.google.com")
	return err == nil
}

// DNSResolver resolves DNS nameservers randomly that previously loaded.
var DnsResolver = &net.Resolver{
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}
		i := int(atomic.AddUint64(&currentResolverIndex, 1)) % len(dnsResolvers)
		return d.DialContext(ctx, "udp", dnsResolvers[i]+":53")
	},
}
