package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	config.TLSConfig.CurvePreferences = []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256}
	config.TLSConfig.CipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}
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
			remoteConnection, err := localListener.Accept()
			if err != nil {
				panic(err)
			}
			fmt.Println("accepted connection from " + remoteConnection.RemoteAddr().String())
			go func(rc net.Conn) {
				localUDPConnection, err := net.Dial("udp4", config.Connect)
				if err != nil {
					panic(err)
				}
				defer localUDPConnection.Close()
				fmt.Printf("%s -> %s -> %s\n", rc.RemoteAddr().String(), localUDPConnection.LocalAddr().String(), config.Connect)
				go func() {
					defer rc.Close()
					buff := make([]byte, 1024*8)
					var n int
					var e error
					for {
						n, e = rc.Read(buff)
						if e != nil {
							fmt.Println(e)
							break
						}
						localUDPConnection.Write(buff[:n])
					}
				}()
				buff := make([]byte, 1024*8)
				var n int
				var e error
				for {
					n, e = localUDPConnection.Read(buff)
					if e != nil {
						fmt.Println(e)
						break
					}
					rc.Write(buff[:n])
				}
			}(remoteConnection)
		}
	} else {
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
		var localClientAddress *net.UDPAddr
		buff := make([]byte, 1024*8)
		var n int
		var connToServer *tls.Conn
		var ok bool
		for {
			n, localClientAddress, _ = localConnection.ReadFromUDP(buff)
			if connToServer, ok = connectionsToServer[localClientAddress.String()]; ok {
				_, err = connToServer.Write(buff[:n])
				if err != nil {
					panic(err)
				}
			} else {
				connToServer, _ = tls.Dial("tcp", config.Connect, &config.TLSConfig)
				connectionsToServer[localClientAddress.String()] = connToServer
				connToServer.Write(buff[:n])
				go func(addr *net.UDPAddr, conn *tls.Conn) {
					buff := make([]byte, 1024*8)
					var n int
					var err error
					for {
						n, _ = conn.Read(buff)
						_, err = localConnection.WriteToUDP(buff[:n], addr)
						if err != nil {
							panic(err)
						}
					}
				}(localClientAddress, connToServer)
				fmt.Printf("accepted connection from %s\n", localClientAddress.String())
			}
		}
	}
}
