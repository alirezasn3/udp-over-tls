package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

var config Config

type Packet struct {
	address *net.UDPAddr
	buffer  []byte
}

type Config struct {
	Role                string `json:"role"`
	Connect             string `json:"connect"`
	Listen              string `json:"listen"`
	CertificateLocation string `json:"certificateLocation"`
	KeyLocation         string `json:"KeyLocation"`
	Threads             int    `json:"threads"`
	TLSConfig           tls.Config
}

func loadConfigFile(config *Config) {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1] + configPath
	}
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
}

func loadCertificates(config *Config) {
	certificate, err := tls.LoadX509KeyPair(config.CertificateLocation, config.KeyLocation)
	if err != nil {
		panic(err)
	}
	config.TLSConfig.MinVersion = tls.VersionTLS13
	config.TLSConfig.Certificates = []tls.Certificate{certificate}
	config.TLSConfig.InsecureSkipVerify = true
}

func main() {
	loadConfigFile(&config)
	loadCertificates(&config)
	if config.Role == "server" {
		localListener, err := tls.Listen("tcp", config.Listen, &config.TLSConfig)
		if err != nil {
			panic(err)
		}
		fmt.Println("listening on " + config.Listen)
		for {
			remoteConnection, _ := localListener.Accept()
			fmt.Println("accepted connection from " + remoteConnection.RemoteAddr().String())
			go func() {
				localUDPConnection, err := net.Dial("udp4", config.Connect)
				if err != nil {
					panic(err)
				}
				fmt.Printf("%s -> %s -> %s\n", remoteConnection.RemoteAddr().String(), localUDPConnection.LocalAddr().String(), config.Connect)
				go func() {
					buff := make([]byte, 1024*8)
					var n int
					for {
						n, _ = remoteConnection.Read(buff)
						localUDPConnection.Write(buff[:n])
					}
				}()
				buff := make([]byte, 1024*8)
				var n int
				for {
					n, _ = localUDPConnection.Read(buff)
					remoteConnection.Write(buff[:n])
				}
			}()
		}
	} else {
		rch := make(chan Packet, 1024)
		connectionsToServer := make(map[string]*tls.Conn)
		listenAddress, err := net.ResolveUDPAddr("udp4", config.Listen)
		if err != nil {
			panic(err)
		}
		localConnection, err := net.ListenUDP("udp4", listenAddress)
		if err != nil {
			panic(err)
		}
		fmt.Println("listening on " + config.Listen)
		for i := 0; i < config.Threads; i++ {
			go func(ch <-chan Packet, j int) {
				var p Packet
				for p = range ch {
					var connToServer *tls.Conn
					var ok bool
					if connToServer, ok = connectionsToServer[p.address.String()]; ok {
						connToServer.Write(p.buffer)
					} else {
						connToServer, _ = tls.Dial("tcp", config.Connect, &config.TLSConfig)
						connectionsToServer[p.address.String()] = connToServer
						connToServer.Write(p.buffer)
						go func(addr *net.UDPAddr, conn *tls.Conn) {
							buff := make([]byte, 1024*8)
							var n int
							for {
								n, _ = conn.Read(buff)
								localConnection.WriteToUDP(buff[:n], addr)
							}
						}(p.address, connToServer)
						fmt.Printf("accepted connection from %s\n", p.address.String())
					}
				}
			}(rch, i)
		}
		for i := 0; i < config.Threads; i++ {
			go func(j int) {
				var localClientAddress *net.UDPAddr
				buff := make([]byte, 1024*8)
				var n int
				for {
					n, localClientAddress, _ = localConnection.ReadFromUDP(buff)
					rch <- Packet{address: localClientAddress, buffer: buff[:n]}
				}
			}(i)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	}
}
