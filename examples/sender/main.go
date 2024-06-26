package main

import (
	"bufio"
	"cf-stun/internal/client"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn.example.com")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	scanner := bufio.NewScanner(os.Stdin)

	var remotePort int
	fmt.Printf("Enter remote port: ")
	if !scanner.Scan() {
		log.Fatalf("Failed to read remote port: %v", scanner.Err())
	}
	remotePort, err = strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("Failed to parse remote port: %v", err)
	}

	relayIP := relayConn.LocalAddr().(*net.UDPAddr).IP
	remoteAddr := &net.UDPAddr{
		IP:   relayIP,
		Port: remotePort,
	}

	log.Println("Enter messages to send to remote")
	for scanner.Scan() {
		msg := scanner.Text()
		_, err := relayConn.WriteTo([]byte(msg), remoteAddr)
		if err != nil {
			log.Printf("Failed to write to relay: %v", err)
		}
	}
}
