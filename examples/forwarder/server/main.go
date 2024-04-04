package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr := flag.String("a", "", "address to be forwarded, e.g. localhost:8080")
	realm := flag.String("realm", "cf-turn-forwarder.example.com", "realm used for TURN")
	flag.Parse()

	if *addr == "" {
		fmt.Println("Please provide an address to be forwarded. Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	turnClient, conn, relayConn, err := client.NewClientConn(*realm)
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())
	log.Printf("Remote port for client: %d", relayConn.LocalAddr().(*net.UDPAddr).Port)

	ln, err := quic.NewListener(relayConn)
	if err != nil {
		log.Panicf("Failed to create server: %v", err)
	}
	defer ln.Close()

	quic.ForwardSessionsAsServer(context.Background(), ln, *addr)
}
