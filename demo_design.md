
⸻

本地MVP 详细设计

“一进程、一房间、四浏览器窗口” 即可完整打一场掼蛋牌局

1. 目标与范围

目标	验收标准
打通 本地 1 房 4 人 全流程	四个浏览器连同一房间，可完成一场 Match；无 panic；操作延迟 < 200 ms
验证 WebSocket 协议 与 快照同步 机制	客户端丢 30% 帧模拟网络抖动，仍能依照 Version 自愈
提供 Docker Compose 一键启动	docker compose up 后：http://localhost:5173 打开即能创建/加入房
覆盖核心单测 ≥ 80 %	go test / vitest 报告达标

超出范围：账号体系、断线重连、数据库、观战、AI Bot 等留待 Phase 2。

⸻

2. 总体架构

┌──────────────────┐    REST    ┌──────────────────┐
│    Web Client    │ ─────────► │   Guandan API    │
│  (React + Vite)  │            │  (Go HTTP/WS)    │
│                  │ <───────── │  ─ GameService ─ │
│  Zustand Store   │   WS/SSE   │  ─ Room Kernel ─ │
└──────────────────┘             └──────────────────┘

	•	单体进程：HTTP(REST) + WS 共用同一 *http.Server。
	•	Room Kernel：封装现有 GameService，房间级锁（sync.Mutex）保证并发安全。
	•	客户端：纯前端静态资源，通过 WebSocket 双向通信，状态由 Zustand 集中管理。

⸻

3. 服务器设计（Go 1.22）

3.1 目录结构

cmd/guandan-server/
├─ main.go                  // 启动、graceful shutdown
├─ handler/                 // HTTP + WS
│  ├─ rest.go               // /createRoom /joinRoom
│  └─ ws.go                 // /room/{id}/ws
├─ room/                    // 单房间内核
│  ├─ kernel.go             // 封装 GameService & players
│  └─ types.go              // PlayerConn, Msg structs
└─ router.go                // Gorilla/Mux 路由

3.2 REST API

Method	Path	Body / Query	返回
POST	/api/room	{ "roomName": "test" }	{ "roomId": "abc123" }
POST	/api/room/{id}/join	{ "seat": 0 }	{ "wsUrl": "ws://…/room/abc123/ws?seat=0" }

3.3 WebSocket 消息协议（JSON）

3.3.1 客户端 → 服务器

// 出牌
{ "t": "PlayCards", "cards": ["♠9","♠9","♠9"] }
// 过
{ "t": "Pass" }

3.3.2 服务器 → 客户端

// 快照推送（全量，仅在大状态变更或 Version 落后时发送）
{ "t": "Snapshot", "version": 42, "payload": { ...MatchSnapshot } }

// 增量事件（日常高频推送）
{ "t": "Event", "e": "CardsPlayed", "data": { "seat": 1, "cards": [...] } }

同步逻辑
	1.	客户端维护 localVersion。
	2.	收到 Snapshot 直接 replaceState(payload)。
	3.	收到 Event 按序 apply；若发现 e.version ≠ localVersion+1，立即请求全量快照。

3.4 核心流程时序

sequenceDiagram
  participant C1 as Client(Seat0)
  participant WS as WS Conn
  participant RK as RoomKernel
  C1->>WS: PlayCards
  WS->>RK: dispatch(seat0,PlayCards)
  RK->>RK: GameService.PlayCards()
  RK-->>WS: broadcast(Event:CardsPlayed)
  WS-->>C1: Event:CardsPlayed


⸻

4. 前端设计（React + Vite + TypeScript）

4.1 目录结构

frontend/
├─ src/
│  ├─ components/
│  │  ├─ Table.tsx       // 中央出牌区
│  │  ├─ Hand.tsx        // 自己的手牌，可拖拽
│  │  ├─ PlayerInfo.tsx  // 头像、座次、余牌数
│  │  └─ ControlBar.tsx  // 出牌 / 过 按钮
│  ├─ pages/
│  │  ├─ Lobby.tsx       // 创建/加入房
│  │  └─ Room.tsx        // 主战场
│  ├─ store.ts           // Zustand + immer
│  ├─ ws.ts              // Promise-style封装
│  └─ types.ts
└─ vite.config.ts

4.2 状态模型（Zustand）

interface RoomState {
  me: { seat: number; hand: Card[] };
  players: Record<SeatID, { handCount: number }>;
  tablePlay: CardGroup | null;
  eventsVersion: number;
  actions: {
    play(cards: Card[]): void;
    pass(): void;
  };
}

4.3 关键交互
	1.	拖拽出牌：使用 react-dnd；合法性客户端先行校验 → WebSocket 发送。
	2.	被动同步：ws.ts 将消息 dispatch 到 store.onWsMsg()；UI 订阅变化自动刷新。
	3.	错误反馈：服务端返回 ErrIllegalPlay → toast 弹窗 + revert 手牌。

4.4 UI 布局草图

┌──────────────────────────────┐
│            Seat2             │
│   [HandCount]   [Avatar]     │
│                              │
│       ┌──── Table ────┐      │
│ Seat3 │  ♠9 ♠9 ♠9     │ Seat1│
│ Info  │              Info   │
│       └────────────────┘     │
│                              │
│  Hand(Seat0)  ControlBar     │
└──────────────────────────────┘


⸻

5. DevOps & 本地运行

5.1 Docker Compose

version: '3'
services:
  server:
    build: ./cmd/guandan-server
    ports: ['8080:8080']
  web:
    build: ./frontend
    environment:
      - VITE_API_HOST=http://localhost:8080
    ports: ['5173:5173']

前端容器 npm run dev -- --host，方便热更新。

5.2 一键启动

docker compose up --build
# 浏览器分别打开 http://localhost:5173
# Lobby 页面中输入同一 RoomId & 不同 Seat 号


⸻

6. 测试计划

层级	工具	用例示例
Go 单测	go test	RoomKernel 并发 500 次 Play/Pass 不丢事件
前端单测	Vitest + React Testing Library	Hand 组件拖拽 & 出牌按钮触发 ws 发送
E2E	Playwright	自动开 4 页浏览器 → 完成一局 Match；断言 UI & 胜负结果


⸻

7. 风险与缓解

风险	影响	缓解方案
WS 消息序列错乱	客户端状态幽灵牌 / 漏牌	引入 Version + 快照自愈
并发写同一房间	数据竞争导致非法状态	RoomKernel 级 mutex；单房间 goroutine 模式可选
浏览器兼容性	拖拽在移动端失效	Phase 1 仅支持桌面；后续引入 Hammer.js


⸻

8. 交付物清单
	•	cmd/guandan-server 源码 + 可执行文件
	•	frontend/ 源码 + 打包产物
	•	docker-compose.yml
	•	docs/phase1-design.md（本文件）
	•	测试报告

⸻
