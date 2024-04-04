package main

import (
	"bufio"
	"cf-stun/internal/client"
	"cf-stun/internal/credentials"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var serverAddr = "turn.speed.cloudflare.com:50000"

func main() {
	creds, err := credentials.Get()
	if err != nil {
		log.Panicf("Failed to get credentials: %v", err)
	}

	cfg := client.ClientConnConfig{
		ServerAddr: serverAddr,
		Username:   creds.Username,
		Password:   creds.Password,
		Realm:      "cf-turn.example.com",
	}

	turnClient, conn, relayConn, err := client.NewClientConn(&cfg)
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	scanner := bufio.NewScanner(os.Stdin)

	var remotePort int
	fmt.Printf("Enter remote port: ")
	if !scanner.Scan() {
		log.Panicf("Failed to read remote port: %v", scanner.Err())
	}
	remotePort, err = strconv.Atoi(scanner.Text())
	if err != nil {
		log.Panicf("Failed to parse remote port: %v", err)
	}

	relayIP := relayConn.LocalAddr().(*net.UDPAddr).IP
	remoteAddr := &net.UDPAddr{
		IP:   relayIP,
		Port: remotePort,
	}

	turnClient.CreatePermission(remoteAddr)

	go func() {
		for {
			buf := make([]byte, 4096)
			n, addr, err := relayConn.ReadFrom(buf)
			if err != nil {
				log.Printf("Failed to read from relay: %v", err)
				continue
			}
			log.Printf("Received from %v: %v", addr, string(buf[:n]))
		}
	}()

	log.Println("Enter messages to send to remote")
	for scanner.Scan() {
		msg := scanner.Text()
		_, err := relayConn.WriteTo([]byte(msg), remoteAddr)
		if err != nil {
			log.Printf("Failed to write to relay: %v", err)
		}
	}
}
