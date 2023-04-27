package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

var config Config

type Config struct {
	Role                string `json:"role"`
	Connect             string `json:"connect"`
	Listen              string `json:"listen"`
	CertificateLocation string `json:"certificateLocation"`
	KeyLocation         string `json:"KeyLocation"`
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
	config.TLSConfig.MinVersion = tls.VersionTLS12
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
				listenAddress, err := net.ResolveUDPAddr("udp", ":0")
				if err != nil {
					panic(err)
				}
				connectAddress, err := net.ResolveUDPAddr("udp", config.Connect)
				if err != nil {
					panic(err)
				}
				localUDPConnection, err := net.ListenUDP("udp", listenAddress)
				if err != nil {
					panic(err)
				}
				fmt.Printf("%s -> %s -> %s", remoteConnection.RemoteAddr().String(), localUDPConnection.LocalAddr().String(), config.Connect)
				go func() {
					buff := make([]byte, 1024*32)
					var n int
					for {
						n, _ = remoteConnection.Read(buff)
						localUDPConnection.WriteToUDP(buff[:n], connectAddress)
					}
				}()
				io.Copy(remoteConnection, localUDPConnection)
			}()
		}
	} else {
		connectionsToServer := make(map[string]*tls.Conn)
		listenAddress, err := net.ResolveUDPAddr("udp", config.Listen)
		if err != nil {
			panic(err)
		}
		localConnection, err := net.ListenUDP("udp", listenAddress)
		if err != nil {
			panic(err)
		}
		fmt.Println("listening on " + config.Listen)
		var localClientAddress *net.UDPAddr
		buff := make([]byte, 1024*32)
		var n int
		for {
			n, localClientAddress, _ = localConnection.ReadFromUDP(buff)
			if connToServer, ok := connectionsToServer[localClientAddress.String()]; ok {
				connToServer.Write(buff[:n])
			} else {
				connectionToServer, _ := tls.Dial("tcp", config.Connect, &config.TLSConfig)
				connectionsToServer[localClientAddress.String()] = connectionToServer
				connectionToServer.Write(buff[:n])
				go func(addr *net.UDPAddr, cs *tls.Conn) {
					buff := make([]byte, 1024*32)
					var n int
					for {
						n, _ = cs.Read(buff)
						localConnection.WriteToUDP(buff[:n], addr)
					}
				}(localClientAddress, connectionToServer)
			}
		}
	}
}
