package cad_test

import (
	"nameserver/cad"
	"testing"
)

func TestAdd(t *testing.T) {
	cad.AddEntry("git.duti.dev", cad.Entry{
		Dest: "127.0.0.1",
		Port: 3000,
		WAF:  true,
	})
	println(cad.GenCaddyfile())

}
