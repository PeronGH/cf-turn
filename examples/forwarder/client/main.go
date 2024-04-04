package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	remotePort := flag.Int("r", 0, "remote port on TURN server")
	localPort := flag.Int("l", 0, "local port to listen on")
	realm := flag.String("realm", "cf-turn-forwarder.example.com", "realm used for TURN")
	flag.Parse()

	if *remotePort == 0 || *localPort == 0 {
		fmt.Println("Please provide remote and local ports. Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	turnClient, conn, relayConn, err := client.NewClientConn(*realm)
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	session, err := quic.NewClientSession(context.Background(), relayConn, *remotePort)
	if err != nil {
		log.Panicf("Failed to create session: %v", err)
	}
	defer func() {
		if err := session.CloseWithError(0, "close"); err != nil {
			log.Printf("session close error: %v", err)
		}
	}()

	quic.ForwardSessionAsClient(context.Background(), session, *localPort)
}
