# 📡 PeerTunnel

PeerTunnel is a minimalist, zero-config, peer-to-peer port forwarding tool based on libp2p. It allows you to expose a local port to another machine without needing a public IP, registering for an account, or setting up a central server.

## Features

*   **No Public IP Required:** Works behind NATs.
*   **Zero-Configuration:** No central server or account registration needed.
*   **NAT Traversal:** Uses libp2p for hole-punching, with fallback to relays.
*   **Simple:** Does one thing well: forwards a TCP port from one machine to another.
*   **Scriptable:** Can be easily integrated into command-line scripts.

## Usage

Suppose you have a server (A) behind a home NAT running an SSH service on port 22, and another server (B).

**On Server A (behind NAT):**

Run the `tcp` command to expose the local port and make it discoverable under a unique tunnel ID.

```bash
pt tcp --port 22 --id ssh.tao.gossip
```

*   `--port 22`: The local port you want to expose.
*   `--id ssh.tao.gossip`: A unique identifier for your tunnel.

**On Server B:**

Run the `bind` command to create a local listener that forwards traffic to Server A.

```bash
pt bind --port 2022 --id ssh.tao.gossip
```

*   `--port 2022`: The local port on Server B that will forward to Server A's port 22.
*   `--id ssh.tao.gossip`: The unique tunnel ID used by Server A.

Now, any traffic sent to `B:2022` will be transparently forwarded to `A:22`.

## Building from Source

To build PeerTunnel, you need Go 1.18 or later.

```bash
# Clone the repository
git clone https://github.com/user/peertunnel.git
cd peertunnel

# Build the executable
CGO_ENABLED=0 go build -o pt
```

## How It Works

PeerTunnel uses [libp2p](https://libp2p.io/) to create a peer-to-peer connection between two nodes.

1.  **Discovery:** The `bind` node advertises itself on the libp2p network under a topic derived from the tunnel ID. The `tcp` node searches for peers on that same topic.
2.  **Connection:** Once the peers discover each other, they attempt to establish a direct connection using NAT traversal techniques (hole-punching). If a direct connection fails, they can use a relay as a fallback.
3.  **Forwarding:** Once a connection is established, PeerTunnel creates a simple, bidirectional data pipe between the local TCP connections and the p2p stream.

## License

This project is licensed under the MIT License.
