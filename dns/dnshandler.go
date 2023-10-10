package dns

import (
	"fmt"
	"log"
	"nameserver/database"
	"os"

	"github.com/miekg/dns"
)

var soaRecord string

func init() {
	nameserver := os.Getenv("NAMESERVER")
	if nameserver == "" {
		nameserver = "localhost"
	}
	soaRecord = fmt.Sprintf("%s. root.%s. 1 604800 86400 2419200 604800", nameserver, nameserver)
}

func Handler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	dbRecords := database.GetDNSRecords(r.Question[0].Name, r.Question[0].Qtype)

	if len(dbRecords) == 0 {
		// Add SOA record
		rr, err := dns.NewRR(fmt.Sprintf("%s IN SOA %s", r.Question[0].Name, soaRecord))
		if err != nil {
			log.Println("Failed to create new RR:", err)
		} else {
			m.Answer = append(m.Answer, rr)
		}
		m.SetRcode(r, dns.RcodeNameError)
	} else {
		for _, dbRecord := range dbRecords {
			rr, err := dns.NewRR(fmt.Sprintf("%s IN %s %s", dbRecord.Domain, dns.TypeToString[dbRecord.RecordType], dbRecord.Value))
			if err != nil {
				log.Println("Failed to create new RR:", err)
				continue
			}
			m.Answer = append(m.Answer, rr)
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Fatal(err.Error())
	}
}
