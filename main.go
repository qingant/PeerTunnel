package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"peertunnel/p2p"
)

const protocolID = "/peertunnel/1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "tcp":
		handleTcpCommand()
	case "bind":
		handleBindCommand()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleTcpCommand() {
	tcpCmd := flag.NewFlagSet("tcp", flag.ExitOnError)
	port := tcpCmd.Int("port", 0, "Local port to expose (required)")
	id := tcpCmd.String("id", "", "Unique tunnel ID (required)")
	tcpCmd.Usage = func() {
		fmt.Println("Usage: pt tcp --port <local-port> --id <tunnel-id>")
		tcpCmd.PrintDefaults()
	}
	tcpCmd.Parse(os.Args[2:])

	if *port == 0 {
		fmt.Println("Error: --port is required")
		tcpCmd.Usage()
		os.Exit(1)
	}
	if *id == "" {
		fmt.Println("Error: --id is required")
		tcpCmd.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := p2p.NewHost(ctx, "")
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	// Subscribe to address change events
	sub, err := h.EventBus().Subscribe(new(event.EvtLocalAddressesUpdated))
	if err != nil {
		log.Fatalf("Failed to subscribe to address change events: %v", err)
	}
	defer sub.Close()

	fmt.Printf("PeerTunnel TCP node started with ID: %s\n", *id)
	fmt.Println("Listening for addresses...")

	// Print initial addresses
	printAddresses(h)

	// Wait for new addresses to be added
	go func() {
		for {
			select {
			case <-sub.Out():
				printAddresses(h)
			case <-ctx.Done():
				return
			}
		}
	}()

	p2p.Advertise(ctx, h, *id)

	h.SetStreamHandler(protocol.ID(protocolID), func(s network.Stream) {
		log.Printf("Received new stream from %s", s.Conn().RemotePeer())
		localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", *port))
		if err != nil {
			log.Printf("Failed to dial local service: %v", err)
			s.Reset()
			return
		}
		log.Printf("Forwarding stream to localhost:%d", *port)
		go p2p.Forwarder(s, localConn)
	})

	// Wait for a SIGINT or SIGTERM signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
}

func handleBindCommand() {
	bindCmd := flag.NewFlagSet("bind", flag.ExitOnError)
	port := bindCmd.Int("port", 0, "Local port to listen on (required)")
	id := bindCmd.String("id", "", "Unique tunnel ID (for discovery)")
	peerAddr := bindCmd.String("peer", "", "Peer multiaddress (bypasses discovery)")
	bindCmd.Usage = func() {
		fmt.Println("Usage: pt bind --port <listen-port> --id <tunnel-id> [--peer <peer-multiaddress>]")
		bindCmd.PrintDefaults()
	}
	bindCmd.Parse(os.Args[2:])

	if *port == 0 {
		fmt.Println("Error: --port is required")
		bindCmd.Usage()
		os.Exit(1)
	}
	if *id == "" && *peerAddr == "" {
		fmt.Println("Error: --id or --peer is required")
		bindCmd.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := p2p.NewHost(ctx, "")
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	var peerInfo *peer.AddrInfo
	if *peerAddr != "" {
		addr, err := multiaddr.NewMultiaddr(*peerAddr)
		if err != nil {
			log.Fatalf("Invalid peer multiaddress: %v", err)
		}
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Fatalf("Failed to create AddrInfo from multiaddress: %v", err)
		}
		peerInfo = pi
	} else {
		fmt.Printf("PeerTunnel Bind node started for ID: %s\n", *id)
		pi, err := p2p.Discover(ctx, h, *id)
		if err != nil {
			log.Fatalf("Failed to discover peer for ID %s: %v", *id, err)
		}
		peerInfo = pi
	}

	log.Printf("Connecting to peer: %s", peerInfo.ID)
	if err := h.Connect(ctx, *peerInfo); err != nil {
		log.Fatalf("Failed to connect to peer: %v", err)
	}
	log.Printf("Connected to peer: %s", peerInfo.ID)

	listenAddr := fmt.Sprintf("localhost:%d", *port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}
	defer listener.Close()
	log.Printf("Listening for local connections on %s", listenAddr)

	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept local connection: %v", err)
			continue
		}
		log.Printf("Accepted local connection from %s", localConn.RemoteAddr())

		s, err := h.NewStream(ctx, peerInfo.ID, protocol.ID(protocolID))
		if err != nil {
			log.Printf("Failed to open stream to peer: %v", err)
			localConn.Close()
			continue
		}
		log.Printf("Opened stream to %s, forwarding traffic", peerInfo.ID)
		go p2p.Forwarder(s, localConn)
	}
}

func printAddresses(h host.Host) {
	fmt.Println("\nUpdated node addresses:")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID())
	}
}

func printUsage() {
	fmt.Println("Usage: pt <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  tcp  --port <local-port> --id <tunnel-id>")
	fmt.Println("  bind --port <listen-port> --id <tunnel-id> [--peer <peer-multiaddress>]")
}
