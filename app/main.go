package main

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/dns-server-starter-go/app/dns"
)

func main() {
	fmt.Println("Starting DNS server on")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()
	buf := make([]byte, 512)
	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}
		receivedData := buf[:size]
		fmt.Println("buf[:size]",buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)
		// Create an empty response
		message := dns.PrepareMessage(&receivedData)
		response := []byte{}
		response = append(response, (*message).Header...)
		response = append(response, (*message).Question...)
		response = append(response, (*message).Answer...)
		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
