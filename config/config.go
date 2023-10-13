package config

import "os"

var ServerIP string
var ServerCNAME string
func init(){
	ServerIP = os.Getenv("SERVER_IP")
	ServerCNAME = os.Getenv("SERVER_CNAME")

	if ServerIP == "" || ServerCNAME == "" {
		panic("SERVER_IP and SERVER_CNAME is required for proxying")	
	}
}
