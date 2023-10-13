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
	}
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if record.Port == 0 {
		record.Port = 80
	}
	var pointsTo string
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
	recordType, ok := dns.StringToType[record.RecordType]
	if !ok {
		c.String(400, "Invalid type")
		return
	}

	if !strings.HasSuffix(record.Domain, ".") {
		record.Domain += "."
	}
	if !(record.Value == pointsTo) {
		entry := &cad.Entry{
			Domain: record.Domain,
			IP:     record.Value,
			Port:   record.Value,
			WAF:    record.WAFEnabled,
		}
		err := database.AddCadEntry(entry)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if err = cad.AddEntry(entry); err != nil {
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
	c.JSON(200, dbRecords)
}

func RemoveRecord(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := database.RemoveDNSRecord(req.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{})
}

func GetDomains(c *gin.Context) {
	c.JSON(200, database.GetDomains())
}
