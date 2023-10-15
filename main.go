package main

import (
	"embed"
	"flag"
	"log"
	"nameserver/api"
	"nameserver/auth"
	"nameserver/cad"
	"nameserver/config"
	"nameserver/database"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/miekg/dns"

	dnshanlder "nameserver/dns"
)

var http_addr string
var dns_addr string

//go:embed static/*
var staticEmbed embed.FS

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
	for _, entry := range entries {
		// Clear domain to save memory
		domain := entry.Domain
		entry.Domain = ""
		cad.AddEntry(domain, entry)
	}
	err = cad.LoadConfig()
	if err != nil {
		log.Println("Failed to load entries")
	}

}

func main() {
	flag.StringVar(&http_addr, "http-addr", "127.0.0.1:8001", "HTTP Listener Address")
	flag.StringVar(&dns_addr, "dns-addr", ":53", "DNS Listener Address")
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	// DNS server
	dns.HandleFunc(".", dnshanlder.Handler)

	go startServer(dns_addr, "tcp")
	log.Println("Listening on ", dns_addr)
	go startServer(dns_addr, "udp")

	// API server
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.GET("/", func(ctx *gin.Context) {
		authCookie, err := ctx.Cookie("auth")
		if err != nil || auth.Validate(authCookie) != nil {
			ctx.Header("Location", "/login.html")
		} else {
			ctx.Header("Location", "/dashboard.html")
		}
		ctx.String(302, "")
	})

	apiGroup := r.Group("/api/")
	apiGroup.Use(authMiddleware)
	apiGroup.POST("/api/records/add", api.AddRecord)
	apiGroup.POST("/api/records/:domain", api.GetRecords)
	apiGroup.POST("/api/records/remove", api.RemoveRecord)
	apiGroup.POST("/api/domains", api.GetDomains)
	r.POST("/auth/login", func(ctx *gin.Context) {
		if pswd, _ := ctx.GetPostForm("password"); pswd == config.ServerPassword {
			token, err := auth.NewToken()
			if err != nil {
				ctx.JSON(401, gin.H{"error": err.Error()})
				return
			}
			ctx.SetCookie("auth", token, 3600, "/", "", true, true)
			ctx.Header("Location", "/dashboard.html")
			ctx.JSON(302, gin.H{})
			return
		}
		ctx.Header("Location", "/login.html?error=true")
		ctx.String(302, "Access denied")
	})
	r.GET("/:staticPath", func(ctx *gin.Context) {
		path := ctx.Param("staticPath")
		if path == "dashboard.html" {
			authMiddleware(ctx)
		}
		file, err := staticEmbed.ReadFile("static/" + path)
		if err != nil {
			ctx.String(404, "Not found")
			return
		}
		pathSplit := strings.Split(path, ".")
		if len(pathSplit) > 1 {
			var contentType string = "text/plain"
			switch pathSplit[len(pathSplit)-1] {
			case "css":
				contentType = "text/css"
			case "js":
				contentType = "text/javascript"
			case "html":
				contentType = "text/html"
			}
			ctx.Header("content-type", contentType)
		}
		ctx.String(200, string(file))
	})

	log.Println("Listening on ", http_addr)
	err := r.Run(http_addr)
	if err != nil {
		panic(err)
	}

}

func authMiddleware(ctx *gin.Context) {
	authCookie, _ := ctx.Cookie("auth")
	if err := auth.Validate(authCookie); err != nil {
		ctx.Header("Location", "/")
		ctx.JSON(302, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}
	ctx.Next()
}
