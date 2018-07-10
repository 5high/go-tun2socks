package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/eycorsican/go-tun2socks/lwip"
	"github.com/eycorsican/go-tun2socks/proxy/socks"
	"github.com/eycorsican/go-tun2socks/tun"
)

func main() {
	tunName := flag.String("tunName", "tun1", "TUN interface name.")
	tunAddr := flag.String("tunAddr", "240.0.0.2", "TUN interface address.")
	tunGw := flag.String("tunGw", "240.0.0.1", "TUN interface gateway.")
	tunMask := flag.String("tunMask", "255.255.255.0", "TUN interface netmask.")
	dnsServer := flag.String("dnsServer", "114.114.114.114,223.5.5.5", "DNS resolvers for TUN interface.")
	proxyType := flag.String("proxyType", "socks", "Proxy handler type.")
	proxyServer := flag.String("proxyServer", "1.1.1.1:1087", "Proxy server address.")

	flag.Parse()

	parts := strings.Split(*proxyServer, ":")
	if len(parts) != 2 {
		log.Fatal("invalid server address")
	}
	proxyAddr := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatal("invalid server port")
	}
	proxyPort := uint16(port)

	// Open the tun device.
	dnsServers := strings.Split(*dnsServer, ",")
	dev, err := tun.OpenTunDevice(*tunName, *tunAddr, *tunGw, *tunMask, dnsServers)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
	}

	// Setup TCP/IP stack.
	lwip.Setup()

	switch *proxyType {
	case "socks":
		// Register TCP and UDP handlers to handle accepted connections.
		lwip.RegisterTCPConnectionHandler(socks.NewTCPHandler(proxyAddr, proxyPort))
		lwip.RegisterUDPConnectionHandler(socks.NewUDPHandler(proxyAddr, proxyPort))
		break
	default:
		log.Fatal("unsupported proxy type")
	}

	// Register an output function to write packets output from lwip stack to tun
	// device, output function should be set before input any packets.
	lwip.RegisterOutputFn(func(data []byte) (int, error) {
		return dev.Write(data)
	})

	// Read packets from tun device and input to lwip stack.
	go func() {
		buf := lwip.NewBytes()
		defer lwip.FreeBytes(buf)
		for {
			n, err := dev.Read(buf[:])
			if err != nil {
				log.Fatal("failed to read from tun device: %v", err)
			}
			err = lwip.Input(buf[:n])
			if err != nil {
				log.Fatal("failed to input data to the stack: %v", err)
			}
		}
	}()

	log.Printf("running tun2socks")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-osSignals
}
