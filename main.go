package main

import (
	"flag"
	"log"
	"nameserver/api"
	"nameserver/cad"
	"nameserver/database"

	"github.com/gin-gonic/gin"

	"github.com/miekg/dns"

	dnshanlder "nameserver/dns"
)

var http_addr string
var dns_addr string
var secret string

func startServer(addr, net string) {
	server := &dns.Server{Addr: addr, Net: net, TsigSecret: nil, ReusePort: true}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func init() {
	// Retrieve entries from database
	entries, err := database.GetEntries()
	if err != nil {
		log.Println("Error! ", err.Error())
		return
	}
	cad.SetEntries(entries)
	err = cad.LoadConfig()
	if err != nil {
		log.Println("Failed to load entries")
	}

}

func main() {
	flag.StringVar(&http_addr, "http-addr", "127.0.0.1:8001", "HTTP Listener Address")
	flag.StringVar(&dns_addr, "dns-addr", ":53", "DNS Listener Address")
	flag.StringVar(&secret, "secret", "", "Authentication secret")
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	// DNS server
	dns.HandleFunc(".", dnshanlder.Handler)

	go startServer(dns_addr, "tcp")
	log.Println("Listening on ", dns_addr)
	go startServer(dns_addr, "udp")

	// API server
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != secret {
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	})
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.POST("/api/records/add", api.AddRecord)
	r.GET("/api/records/:domain", api.GetRecords)
	r.POST("/api/records/remove", api.RemoveRecord)
	r.GET("/api/domains", api.GetDomains)

	log.Println("Listening on ", http_addr)
	err := r.Run(http_addr)
	if err != nil {
		panic(err)
	}

}
