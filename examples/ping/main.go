package main

import (
	"cf-stun/internal/client"
	"log"
	"time"
)

func main() {
	turnClient, conn, relayConn, err := client.NewClientConn("cf-turn.example.com")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.CloseClientConn(turnClient, conn, relayConn)
	log.Printf("TURN Client: %v", relayConn.LocalAddr())

	// Start pinging
	buf := make([]byte, 1024)
	data := []byte("ping")

	count := 10
	times := make([]time.Duration, 0, count)

	for range count {
		startTime := time.Now()
		_, err = relayConn.WriteTo(data, relayConn.LocalAddr())
		if err != nil {
			log.Fatalf("Failed to write: %v", err)
		}

		_, _, err = relayConn.ReadFrom(buf)
		if err != nil {
			log.Fatalf("Failed to read: %v", err)
		}

		log.Printf("Ping: %v", time.Since(startTime))
		times = append(times, time.Since(startTime))
	}

	// Calculate min, max, avg
	min := times[0]
	max := times[0]
	avg := times[0]
	for _, t := range times {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		avg += t
	}
	avg /= time.Duration(count)

	log.Printf("Min: %v, Max: %v, Avg: %v", min, max, avg)
}
