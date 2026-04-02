package main

import (
	"log"
	"net"
	"os/exec"
	"os/signal"
	// "strconv"

	"os"

	// "github.com/joho/godotenv"
	"github.com/songgao/water"
)


func main() {

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// portString := os.Getenv("PORT")
	// port, err := strconv.ParseInt(portString, 10, 64)
    ifce, err := water.New(water.Config{DeviceType: water.TUN})

	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("ip", "addr", "add", "10.1.0.20", "peer", "10.1.0.10", "dev", ifce.Name())
	if err := cmd.Run(); err != nil {
		log.Fatalf("Ошибка назначения IP на сервере: %v", err)
	}

	cmd = exec.Command("ip", "link", "set", "dev", ifce.Name(), "up")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Ошибка поднятия интерфейса на сервере: %v", err)
	}
    lst, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 9999})

	var clientAddr *net.UDPAddr 
	go func () {
		packet := make([]byte, 2000)
		for {
			n, addr, err := lst.ReadFromUDP(packet)
			if err != nil {
				continue
			}
			ifce.Write(packet[:n])
			log.Printf("Получено от %s и записано в TUN: %d байт", addr, n)
			clientAddr = addr
		}
	}()

	go func () {
		buf := make([]byte, 2000)
		for {
			n, err := ifce.Read(buf)
			if err != nil {
				continue
			}
			if clientAddr == nil {
            	continue 
        	}
			n, err = lst.WriteToUDP(buf[:n], clientAddr)
			if err != nil {
				continue
			}
			log.Printf("Клиенту отправлено %d байт", n)
		}
	}()

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, os.Interrupt)
	<- osChan
	
}