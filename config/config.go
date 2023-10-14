package config

import "os"

var ServerIP string
var ServerCNAME string
var ServerPassword string

func init(){
	ServerIP = os.Getenv("SERVER_IP")
	ServerCNAME = os.Getenv("SERVER_CNAME")
	ServerPassword = os.Getenv("SERVER_PASSWORD")

	if ServerIP == "" || ServerCNAME == "" {
		panic("SERVER_IP and SERVER_CNAME is required for proxying")	
	}
}
