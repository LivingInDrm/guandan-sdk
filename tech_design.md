掼蛋游戏 SDK —— MVP 设计 & 技术规范

本文是掼蛋游戏SDK的技术设计，后续将其对接 Web 前端或 CLI，来让用户进行游戏。

⸻

1. 顶层目录

sdk/
├─ domain/          # 静态规则 & 纯数据
├─ engine/          # 状态机 + 出牌循环
├─ service/         # 对外 API
└─ event/           # 领域事件总线


⸻

2. 各层职责

层	只做一件事
domain	牌对象、牌型解析、牌力比较、Level→Trump 映射… 全部无状态、纯函数
engine	Match / Deal / Trick 状态迁移；执行 play / pass；抛领域事件
service	场景用例：createMatch  playCards  pass  getSnapshot；线程安全；幂等校验
event	进程内 chan DomainEvent；外层可订阅做 UI 更新 / AI 决策 / 持久化


⸻

3. 领域模型（domain/）

实体	说明
Card	花色 + 点数；大小王用花色 JOKER
CardGroup	一手牌 + cmpKey=(CAT,SIZE,RANK)
Deck	108 张；shuffle(seed) 支持回放
Seat / Player / Team	座次、玩家、阵营
MatchCtx / DealCtx / TrickCtx	局级 / 盘级 / 墩级上下文，皆不可变
枚举	LevelEnum  CatEnum  CmpResult …

⚙️  牌型比较：compare(a,b)、canBeat(hand,tablePlay) 全在 domain。

⸻

4. 引擎（engine/）

4.1 DealStateMachine（P0–P6）

[*] → P1:createMatch → P2:发牌  
P2 → P3(非首Deal) → P4:Tribute → P5:首出 → P6:rankList  
P6 → P1(未到A) / [*](到A)

	•	engine.playCards(seat,cards)
	1.	调 domain.canBeat
	2.	更新 ctx → 产生 CardsPlayed 事件
	•	engine.pass(seat) → 更新 passSet → 产生 PlayerPassed

4.2 事件流

MatchCreated → DealStarted → TributeRequested → … → DealEnded → MatchEnded


⸻

5. Service API（Go 示意）

type MatchOptions struct {
    DealLimit int  // 可选
    Seed      int  // 重放
}

type GameService interface {
    CreateMatch(players []Player, opt *MatchOptions) MatchID
    StartNextDeal(mid MatchID)
    PlayCards(mid MatchID, seat SeatID, cardIDs []CardID)
    Pass(mid MatchID, seat SeatID)
    GetSnapshot(mid MatchID) MatchSnapshot
    Subscribe(mid MatchID, cb func(DomainEvent)) (unsub func())
}

	•	非法动作统一返回 errs.ErrIllegalPlay。
	•	MVP 内部用 sync.Mutex 包级锁保证线程安全。

⸻

6. 快照 & 重放

type MatchSnapshot struct {
    Version     int
    MatchCtx    …
    DealCtx     …
    TrickCtx    …
    Hands       map[SeatID][]Card
    RankList    []SeatID
    History     []DomainEvent
}

	•	序列化：encoding/json
	•	重放：engine.Replay(events) → 单元测试 / 回放 / Debug

⸻

7. MVP 技术选型

范畴	选型	备注
语言	Go 1.22	单文件部署，后续可编译 Wasm
序列化	encoding/json	调试友好
事件总线	chan DomainEvent	后续可换 NATS/Kafka
并发	sync.Mutex	简单可靠；未来可升级 Actor
日志	标准库 log	零依赖
测试	go test	覆盖率 ≥ 80 %
CI	GitHub Actions：go vet + go test	单步骤


⸻

8. 最少测试清单
	1.	CardCompareTest：列举牌型对并断言大小
	2.	DealFlowTest：固定 seed 跑完一局，验最终 Level
	3.	ConcurrencyTest：多 goroutine 并发 play/pass 不丢事件

⸻

9. 部署示例

cmd/guandan-server/
├─ main.go      // HTTP + WebSocket demo
└─ router.go

未来正式线上可拆分 match-room 服务 + API 网关，但 MVP 保持单体。



