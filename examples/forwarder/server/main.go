package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"log"
)

func main() {
	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn-forwarder.example.com")
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	ln, err := quic.NewListener(relayConn)
	if err != nil {
		log.Panicf("Failed to create server: %v", err)
	}
	defer ln.Close()

	quic.ForwardSessionsAsServer(context.Background(), ln, "localhost:3000")
}
