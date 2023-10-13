package cad

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Entry struct {
	Domain string `gorm:"uniqueIndex"` // Private. For database use
	IP     string
	Port   int16
	WAF    bool
}

var entries map[string]Entry // Map of domain name to IP/Port/WAF

func AddEntry(domain string, e Entry) {
	entries[domain] = e
}

func RemoveEntry(domain string) {
	delete(entries, domain)
}

func GetEntry(domain string) *Entry {
	entry := entries[domain]
	return &entry
}

func construct(domain, ip string, port int16, wafEnabled bool) string {
	var tls string
	if domain == "localhost" {
		tls = "tls internal"
	}
	return "\n" + fmt.Sprintf(`https://%s {
		ja3 block_bots %t
		%s
		reverse_proxy http://%s:%d
}`, domain, wafEnabled, tls, ip, port)
}

func GenCaddyfile() string {
	caddyfile := `
{
	order ja3 before respond
	servers {
		listener_wrappers {
			http_redirect
			ja3
			tls
		}
	}
}
	`
	for domain, entry := range entries {
		caddyfile += construct(domain, entry.IP, entry.Port, entry.WAF)
	}
	log.Println(caddyfile)
	return caddyfile
}

func LoadConfig() error {
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:2019/load", strings.NewReader(GenCaddyfile()))
	req.Header.Add("content-type", "text/caddyfile")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("failed to update config")
	}
	defer resp.Body.Close()
	return nil
}
