package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go-ping <hostname>")
		os.Exit(1)
	}

	host := os.Args[1]
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		fmt.Println("Error resolving host:", err)
		os.Exit(1)
	}

	fmt.Printf("PING %s (%s):\n", host, addr.IP.String())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	icmpConn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("Error creating ICMP connection:", err)
		os.Exit(1)
	}
	defer icmpConn.Close()

	seq := 1
	for {
		select {
		case <-interrupt:
			fmt.Println("\nPing utility terminated.")
			return
		default:
			sendPing(icmpConn, addr, seq)
			seq++
			time.Sleep(time.Second)
		}
	}
}

func sendPing(conn net.PacketConn, addr *net.IPAddr, seq int) {
	msg := make([]byte, 8)
	msg[0] = 8 // ICMP Echo Request Type
	msg[1] = 0 // Code
	msg[2] = 0 // Checksum (initially 0)
	msg[3] = 0 // Checksum (initially 0)
	msg[4] = byte(seq >> 8)
	msg[5] = byte(seq)
	msg[6] = 0 // Data (High byte)
	msg[7] = 0 // Data (Low byte)

	checksum := calculateChecksum(msg)
	msg[2] = byte(checksum >> 8)
	msg[3] = byte(checksum)

	if _, err := conn.WriteTo(msg, addr); err != nil {
		fmt.Println("Error sending ICMP packet:", err)
	}
}

func calculateChecksum(data []byte) uint16 {
	sum := 0
	for i := 0; i < len(data)-1; i += 2 {
		sum += int(data[i])<<8 | int(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += int(data[len(data)-1]) << 8
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)
	return uint16(^sum)
}
