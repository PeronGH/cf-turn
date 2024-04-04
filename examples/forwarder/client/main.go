package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"fmt"
	"log"
)

func main() {
	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn-forwarder.example.com")
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	var remotePort int
	fmt.Printf("Enter remote port: ")
	fmt.Scanf("%d\n", &remotePort)

	session, err := quic.NewClientSession(context.Background(), relayConn, remotePort)
	if err != nil {
		log.Panicf("Failed to create session: %v", err)
	}
	defer func() {
		if err := session.CloseWithError(0, "close"); err != nil {
			log.Printf("session close error: %v", err)
		}
	}()

	quic.ForwardSessionAsClient(context.Background(), session, 8000)
}
