package main

import (
	"cf-stun/internal/client"
	"log"
)

func main() {
	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn.example.com")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	for {
		buf := make([]byte, 4096)
		n, addr, err := relayConn.ReadFrom(buf)
		if err != nil {
			log.Printf("Failed to read from relay: %v", err)
			continue
		}
		log.Printf("Received from %v: %v", addr, string(buf[:n]))
	}
}
