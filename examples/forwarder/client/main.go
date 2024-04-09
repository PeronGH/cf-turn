package main

import (
	"cf-stun/internal/client"
	"cf-stun/internal/quic"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/charmbracelet/log"
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
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Infof("TURN Client: %v", relayConn.LocalAddr())

	ctx, cancel := context.WithCancel(context.Background())

	// Establish QUIC session
	session, err := quic.NewClientSession(ctx, relayConn, *remotePort)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer func() {
		cancel()
		if err := session.CloseWithError(0, "close"); err != nil {
			log.Warnf("session close error: %v", err)
		}
	}()

	// Register signal handler to cancel context
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		cancel()
	}()

	// Forward session as client
	quic.ForwardSessionAsClient(ctx, session, *localPort)
}
