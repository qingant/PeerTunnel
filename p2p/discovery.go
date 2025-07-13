package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const protocolPrefix = "/peertunnel/1.0.0/"

// Advertise makes the host discoverable on a given topic.
func Advertise(ctx context.Context, h host.Host, topic string) {
	kdht, err := dht.New(ctx, h)
	if err != nil {
		log.Printf("Failed to create DHT: %v", err)
		return
	}

	if err = bootstrapDHT(ctx, h, kdht); err != nil {
		log.Printf("Failed to bootstrap DHT: %v", err)
		return
	}

	routingDiscovery := routing.NewRoutingDiscovery(kdht)
	util.Advertise(ctx, routingDiscovery, protocolPrefix+topic)
	fmt.Printf("Successfully advertised topic: %s\n", topic)
}

// Discover discovers a peer on a given topic and returns its AddrInfo.
func Discover(ctx context.Context, h host.Host, topic string) (*peer.AddrInfo, error) {
	kdht, err := dht.New(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if err = bootstrapDHT(ctx, h, kdht); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	routingDiscovery := routing.NewRoutingDiscovery(kdht)

	fmt.Printf("Searching for peer with ID: %s\n", topic)
	peerChan, err := routingDiscovery.FindPeers(ctx, protocolPrefix+topic)
	if err != nil {
		return nil, fmt.Errorf("failed to find peers: %w", err)
	}

	// Block until a peer is found
	select {
	case peerInfo := <-peerChan:
		if peerInfo.ID != "" {
			return &peerInfo, nil
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second): // Timeout for discovery
		return nil, fmt.Errorf("timed out waiting for peer discovery")
	}
	return nil, fmt.Errorf("no peer found")
}

func bootstrapDHT(ctx context.Context, h host.Host, kdht *dht.IpfsDHT) error {
	// Connect to bootstrap peers
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				// log.Printf("Error while connecting to bootstrap peer: %s\n", err)
			}
		}()
	}
	wg.Wait()

	return kdht.Bootstrap(ctx)
}
