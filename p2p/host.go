package p2p

import (
	"context"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

var defaultBootstrapPeers = []string{
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ", // mars.i.ipfs.io
}

// NewHost creates a new libp2p host.
func NewHost(ctx context.Context, relays string) (host.Host, error) {
	// Combine default and custom bootstrap peers
	allBootstrapAddrs := defaultBootstrapPeers
	if relays != "" {
		allBootstrapAddrs = append(allBootstrapAddrs, strings.Split(relays, ",")...)
	}

	var bootstrapPeers []peer.AddrInfo
	for _, pAddr := range allBootstrapAddrs {
		maddr, err := multiaddr.NewMultiaddr(pAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap peer address %s: %w", pAddr, err)
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create AddrInfo from multiaddress: %w", err)
		}
		bootstrapPeers = append(bootstrapPeers, *addrInfo)
	}

	var newDHT routing.PeerRouting
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0", // Listen on all IPv4 addresses
			"/ip6/::/tcp/0",      // Listen on all IPv6 addresses
		),
		libp2p.Transport(ws.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.EnableRelay(),
		libp2p.EnableAutoNATv2(),
		libp2p.EnableRelayService(),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			newDHT, err = dht.New(ctx, h, dht.BootstrapPeers(bootstrapPeers...))
			return newDHT, err
		}),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}
