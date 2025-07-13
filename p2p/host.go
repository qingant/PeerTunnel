package p2p

import (
	"context"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

var defaultRelays = []string{
	"/dns4/libp2p-relay.cloudflare-ipfs.com/tcp/443/wss/p2p/12D3KooWDpZXDw5BMbF9zKZ48DYoA9Y1SRMxvZ4ZYmEnQ2WkAkgh",
	"/dns4/libp2p-relay1.ipfs.io/tcp/443/wss/p2p/12D3KooWH9P5VGUUNzH6oCN8HDJqSukTzP3U5sM4PK9Q5z3w9w67",
}

// NewHost creates a new libp2p host.
func NewHost(ctx context.Context, relays string) (host.Host, error) {
	// Combine default and custom relays
	allRelayAddrs := defaultRelays
	if relays != "" {
		allRelayAddrs = append(allRelayAddrs, strings.Split(relays, ",")...)
	}

	var relayAddrInfos []peer.AddrInfo
	for _, rAddr := range allRelayAddrs {
		maddr, err := multiaddr.NewMultiaddr(rAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid relay address %s: %w", rAddr, err)
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create AddrInfo from multiaddress: %w", err)
		}
		relayAddrInfos = append(relayAddrInfos, *addrInfo)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0", // Listen on all IPv4 addresses
			"/ip6/::/tcp/0",      // Listen on all IPv6 addresses
		),
		libp2p.Transport(ws.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.EnableRelay(),
		libp2p.EnableAutoNATv2(),
		libp2p.EnableAutoRelayWithStaticRelays(relayAddrInfos),
		libp2p.Security(noise.ID, noise.New),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}
