package cad

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Entry struct {
	Domain string
	IP     string
	Port   string
	WAF    bool
}

var entries []*Entry

func AddEntry(e *Entry) error {
	entries = append(entries, e)
	return LoadConfig()

}

func RemoveEntry(domain string) error {
	for idx, entry := range entries {
		if entry.Domain == domain {
			entries = append(entries[:idx], entries[idx+1:]...)
		}
	}
	return LoadConfig()
}

func construct(domain, ip, port string, wafEnabled bool) string {
	return "\n" + fmt.Sprintf(`https://%s {
		ja3 block_bots %t
		reverse_proxy http://%s:%s
	}`, domain, wafEnabled, ip, port)
}

func regen() string {
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
	for _, entry := range entries {
		caddyfile += construct(entry.Domain, entry.IP, entry.Port, entry.WAF)
	}
	return caddyfile
}

func LoadConfig()error{
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:2019", strings.NewReader(regen()))
	req.Header.Add("content-type","text/caddyfile")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("failed to update config")
	}
	return nil
}
