package cad_test

import (
	"testing"
	"nameserver/cad"
)

func TestAdd(t *testing.T) {
	cad.AddEntry(&cad.Entry{
		Domain: "git.duti.dev",
		IP: "127.0.0.1",
		Port: "3000",
		WAF: true,
	})
	println(cad.GenCaddyfile())
	
}
