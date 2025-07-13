📡 PeerTunnel: 一个基于 libp2p 的点对点端口转发 / 内网穿透工具

🧠 项目目标

实现一个极简、免注册、开箱即用的内网穿透工具，支持点对点 TCP 端口转发，目标是替代复杂的 VPN 工具和中心化的 FRP 服务。
	•	✅ 不需要公网 IP
	•	✅ 不需要注册账户或配置中心服务器
	•	✅ 默认通过 NAT 穿透（基于 libp2p），无法穿透时回落使用 relay
	•	✅ 只做一件事：将一个机器的端口暴露到另一个机器上
	•	✅ 可嵌入命令行 / 脚本中使用

⸻

📌 使用示例

假设 Server A 在家庭 NAT 后，运行了一个 ssh 服务在 22 端口；Server B 是一个公网 VPS。

在 A 上运行：

pt tcp 22 --id ssh.tao.gossip --relays=wss://relay.nostr.org

在 B 上运行：

pt bind ssh.tao.gossip --port 2022

此时，访问 B:2022 就等价于访问 A:22，实现端口转发，穿透 NAT。

⸻

🏗️ 技术架构

1. 通信协调与标识分发（信令层）
	•	使用 Nostr 协议（或简化版）作为信令通道，进行 peer ID 与连接意图的交换
	•	使用 gossip topic 实现 ID 广播与 relay 发现（可选）
	•	客户端发出连接意图（携带自身 peer ID），绑定端监听后回应连接准备

2. 数据传输层
	•	使用 libp2p 作为连接建立框架
	•	支持 NAT 穿透（hole punching）
	•	支持 TCP、WebRTC、QUIC
	•	fallback 到 libp2p relay 节点以确保连接成功率

3. 数据转发机制
	•	一旦连接建立：
	•	A 端打开本地 TCP 连接（如 localhost:22）
	•	B 端打开监听端口（如 B:2022）
	•	通过 libp2p 连接，将两端连接双向转发

  [Client] <--TCP--> A:22
                  |
                  | libp2p stream
                  ↓
  [Server] <--TCP--> B:2022



⸻

✅ 最小可行版本（MVP）
	•	✅ 单一 relay 下连接
	•	✅ 支持 TCP 映射
	•	✅ 客户端与服务端通过一个预定义 ID 建立连接
	•	✅ 连接成功后可实现完整 TCP 双向数据转发

🚀 开发建议语言

Go
