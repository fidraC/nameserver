package database

import (
	"github.com/glebarez/sqlite"
	"github.com/miekg/dns"
	"gorm.io/gorm"
)

type DNSRecord struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Domain     string `gorm:"index" json:"domain"`
	RecordType uint16 `json:"type"`
	Value      string `json:"value"`
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("dnsRecords.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&DNSRecord{})
}


func GetDNSRecords(domain string, recordType uint16) []DNSRecord {
	var dbRecords []DNSRecord
	cond := "domain = ?"
	args := []interface{}{domain}

	if recordType == 0 {
		// No additional condition
	} else if recordType == dns.TypeA || recordType == dns.TypeAAAA {
		cond += " AND (record_type = ? OR record_type = ?)"
		args = append(args, recordType, dns.TypeCNAME)
	} else {
		cond += " AND record_type = ?"
		args = append(args, recordType)
	}

	err := db.Where(cond, args...).Find(&dbRecords).Error
	if err != nil {
		return nil
	}
	return dbRecords
}


func AddDNSRecord(domain string, recordType uint16, value string) error {
	record := DNSRecord{Domain: domain, RecordType: recordType, Value: value}
	return db.Create(&record).Error
}

func RemoveDNSRecord(id uint) error {
	return db.Delete(&DNSRecord{}, id).Error
}

func GetDomains() []string {
	var domains []string
	err := db.Model(&DNSRecord{}).Distinct().Pluck("domain", &domains).Error
	if err != nil {
		return nil
	}
	return domains
}
