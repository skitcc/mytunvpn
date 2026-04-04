package main

import (
	"log"
	"net"
	"os/exec"
	"os/signal"
	"sync"
	"os"
	"github.com/songgao/water"
)


func main() {
    ifce, err := water.New(water.Config{DeviceType: water.TUN})

	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("sudo", "ip", "addr", "add", "10.1.0.1/24", "dev", ifce.Name())
	if err := cmd.Run(); err != nil {
		log.Fatalf("Ошибка назначения IP на сервере: %v", err)
	}

	cmd = exec.Command("ip", "link", "set", "dev", ifce.Name(), "up")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Ошибка поднятия интерфейса на сервере: %v", err)
	}
    lst, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 9999})

	var clientMap = make(map[string]*net.UDPAddr)
	mutex := &sync.RWMutex{}
	go func () {
		packet := make([]byte, 2000)
		for {
			n, addr, err := lst.ReadFromUDP(packet)
			if err != nil {
				continue
			}
			ifce.Write(packet[:n])
			log.Printf("Получено от %s и записано в TUN: %d байт", addr, n)
			version := packet[0] >> 4
			var srcIp string
			switch version {
			case 4:
				srcIp = net.IP(packet[12:16]).String()
			// case 6:
			// 	srcIp = net.IP(packet[8:24]).String()
			}
			mutex.Lock()
			clientMap[srcIp] = addr
			mutex.Unlock()			
		}
	}()

	go func () {
		buf := make([]byte, 2000)
		for {
			n, err := ifce.Read(buf)
			if err != nil {
				continue
			}
			version := buf[0] >> 4
			if version == 4 {
				dstIp := net.IP(buf[16:20]).String()

				mutex.RLock()
				targetAddr, exists := clientMap[dstIp]
				mutex.RUnlock()

				if exists {
					_, err = lst.WriteToUDP(buf[:n], targetAddr)
					if err != nil {
						log.Printf("Ошибка отправки клиенту %s: %v", dstIp, err)
					} else {
						log.Printf("Отправлено ответ клиенту %s (%s)", dstIp, targetAddr)
					}
				}
			}
		}
	}()

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, os.Interrupt)
	<- osChan
	
}