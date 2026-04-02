package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/songgao/water"
)

func writePacket(ifce *water.Interface, conn *net.UDPConn) {
	for {
		packet := make([]byte, 2000)
		n, err := ifce.Read(packet)
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("Packet Received: % x\n", packet[:n])
		version := packet[0] >> 4
		log.Println(version)
		switch version {
		case 4:
			srcIP := net.IP(packet[12:16])
			dstIP := net.IP(packet[16:20])
			log.Printf("Пакет: %s -> %s", srcIP, dstIP)
			log.Printf("version: %d" ,version)
		case 6:
			// srcIP := net.IP(packet[8:24])
			// dstIP := net.IP(packet[24:40])
			// log.Printf("Пакет: %s -> %s", srcIP, dstIP)
			// log.Printf("version: %d" ,version)
			continue
		};

		_, err = conn.Write(packet[:n])
		if err != nil {
			log.Fatal(err)
		}
	}
}

func readPacket(ifce *water.Interface, conn *net.UDPConn) {
	packet := make([]byte, 2000)
	for {
		n, err := conn.Read(packet)
		log.Printf("Считано %d байт", n)
		n, err = ifce.Write(packet[:n])
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	ip := os.Getenv("IP")
	
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Interface Name: %s\n", ifce.Name())

	cmd := exec.Command("sudo", "ip", "addr", "add", "10.1.0.10", "peer", "10.1.0.20", "dev", ifce.Name())
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("sudo", "ip", "link", "set", "dev", ifce.Name(), "up")
	cmd.Run()
	servaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		log.Fatalf("Error : %v", err)
	}
	conn, _ := net.DialUDP("udp", nil, servaddr)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)
	go writePacket(ifce, conn)
	go readPacket(ifce, conn)
	<-osSignal

}

