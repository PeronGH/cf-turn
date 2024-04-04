package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"flag"
	"log"
)

func main() {
	remotePort := flag.Int("r", 0, "remote port to forward to")
	localPort := flag.Int("l", 0, "local port to listen on")
	flag.Parse()

	if *remotePort == 0 {
		log.Fatal("Remote port to forward to is required")
	}
	if *localPort == 0 {
		log.Fatal("Local port to listen on is required")
	}

	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn-forwarder.example.com")
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
