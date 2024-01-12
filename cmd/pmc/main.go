package main

import (
	"flag"
	"log"
	"net"

	"github.com/xdavidwu/password-miner/internal/stratum"
)

func main() {
	addr := flag.String("address", "127.0.0.1:1234", "Address of the pool")
	username := flag.String("username", "username", "Username")
	password := flag.String("password", "password", "Password")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	for {
		conn, err := net.Dial("tcp", *addr)
		if err != nil {
			log.Fatalf("Cannot connect to pool: %v", err)
		}

		stratum.Stratum(conn, conn, stratum.StratumLoginParams{
			Login: *username,
			Pass: *password,
		})

		log.Printf("Connection died, retrying")
	}
}
