package api

import (
	"nameserver/cad"
	"nameserver/config"
	"nameserver/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

func AddRecord(c *gin.Context) {
	var record struct {
		Domain     string `json:"domain" binding:"required"`
		RecordType string `json:"type" binding:"required"`
		Value      string `json:"value" binding:"required"`
		Port       int16  `json:"port"`
		WAFEnabled bool   `json:"waf_enabled"`
		Proxy      bool   `json:"proxy"`
	}
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	recordType, ok := dns.StringToType[record.RecordType]
	// Before doing anything, check if we're actually proxying. If not, just add DNS record
	if !record.Proxy {
		err := database.AddDNSRecord(record.Domain, recordType, record.Value)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
		return
	}

	// Go's default int value is 0. Should be set to 80 instead
	if record.Port == 0 {
		record.Port = 80
	}
	var pointsTo string

	// AAAA isn't supported yet. We'll just point to our IPv4 and proxy via IPv6
	switch record.RecordType {
	case "A":
		pointsTo = config.ServerIP
	case "AAAA":
		{
			record.RecordType = "A"
			pointsTo = config.ServerIP
		}
	case "CNAME":
		pointsTo = config.ServerCNAME
	default:
		pointsTo = record.Value
	}

	if !ok {
		c.JSON(400, gin.H{"error": "Invalid type"})
		return
	}
	// This checks if record type is not A/AAAA/CNAME. This means we can't proxy it (no need for caddy)
	if !(record.Value == pointsTo) {
		entry := cad.Entry{
			IP:   record.Value,
			Port: record.Port,
			WAF:  record.WAFEnabled,
		}

		cad.AddEntry(record.Domain, entry)
		if err := cad.LoadConfig(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		// Assign domain for database only.
		entry.Domain = record.Domain
		err := database.AddCadEntry(&entry)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	err := database.AddDNSRecord(record.Domain, recordType, pointsTo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{})
}

type record struct {
	Cad *cad.Entry
	DNS *database.DNSRecord
}

func GetRecords(c *gin.Context) {
	domain := c.Param("domain")

	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	dbRecords := database.GetDNSRecords(domain, 0)
	if len(dbRecords) == 0 {
		c.JSON(404, gin.H{})
		return
	}
	var records []record
	for _, dnsRecord := range dbRecords {
		record := record{
			DNS: &dnsRecord,
		}
		if dnsRecord.Value == config.ServerIP || dnsRecord.Value == config.ServerCNAME {
			record.Cad = cad.GetEntry(strings.TrimSuffix(domain, "."))
		}
		records = append(records, record)
	}
	c.JSON(200, records)
}

func RemoveRecord(c *gin.Context) {
	// We need both the ID for when there are multiple entries for DNS and the domain since the ID isn't linked to caddy
	var req struct {
		ID     *uint  `json:"id"`
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Domain != "" {
		cad.RemoveEntry(req.Domain)
		if err := cad.LoadConfig(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	if req.ID != nil {
		err := database.RemoveDNSRecord(*req.ID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(200, gin.H{})
}

func GetDomains(c *gin.Context) {
	c.JSON(200, database.GetDomains())
}
