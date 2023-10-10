package api

import (
	"nameserver/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

func AddRecord(c *gin.Context) {
	var record struct {
		Domain     string `json:"domain"`
		RecordType string `json:"type"`
		Value      string `json:"value"`
	}
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	recordType, ok := dns.StringToType[record.RecordType]
	if !ok {
		c.String(400, "Invalid type")
		return
	}

	if !strings.HasSuffix(record.Domain, ".") {
		record.Domain += "."
	}

	err := database.AddDNSRecord(record.Domain, recordType, record.Value)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
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
