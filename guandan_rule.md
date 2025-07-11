## 掼蛋总体规则

1. **阵营与牌数**
   * 四人对坐成两队，打两副牌（108 张），每人27张牌
2. **目标**
   * 每一场(Match)比赛包含多局(Deal)牌局,每一局牌局的获胜者将获得一定的牌级提升
   * 队伍牌级从 **2** 升到 **A**；率先达到**A**的队伍即可赢整场。

## 关键词速查

| 关键词                  | 中文译名   | 详细含义                                                                                        |
| -------------------- | ------ | ------------------------------------------------------------------------------------------- |
| **Deal**             | 局 / 轮次 | 一整轮游戏流程：从 108 张牌的洗牌、发牌开始，到所有玩家出完手牌并完成结算为止。                                                  |
| **Trick**            | 墩      | 单次出牌序列：由领牌者先出牌，随后顺时针跟牌，直至除最近一次出牌者外的所有 *active players* 连续 PASS 为止；Trick 赢家获得下一个 Trick 的领牌权。 |
| **Hand**             | 手牌     | 某玩家当前持有的全部牌张；随着出牌而减少。                                                                       |
| **Level**            | 牌级     | 队伍升级进度（2 → A）；决定本 Deal 的 Trump 点数。                                                          |
| **Trump**            | 主牌     | 点数 = 当前牌级的 8 张牌；牌力仅次于 **大王 / 小王**。                                              |
| **Big Joker**        | 大王     | 牌力最高的单张牌。                                                                                   |
| **Small Joker**      | 小王     | 牌力次高的单张牌，仅小于大王。                                                                             |
| **Starting Card**    | 首出牌    | 首 Deal 切牌决定的卡牌；持有者成为本局首位领牌者。                                                                |
| **Tribute**          | 上贡     | 上一 Deal 末游方需交出的 **最高非主牌单张**（贡牌）给赢家队伍。                                                       |
| **Return Tribute**   | 还贡     | 赢家队 Seat 1 在选择贡牌后返还给原持有者的一张 ≤ 10 点数的任意单张。                                                   |
| **Tribute Immunity** | 抗贡     | 当末游方（或其两名队员）同时握有两张大王时触发；免除 **上贡 / 还贡** 流程。                                                  |
| **Double Down**      | 双下     | 结算名次 = 1st & 2nd（同队） + 3rd & 4th（另一队）；赢家队牌级 +3。                                             |
| **Single Last**      | 单末游    | 结算名次 = 1st & 3rd（同队） + 2nd & 4th（另一队）；赢家队牌级 +2。                                             |
| **Partner Last**     | 自家末游   | 结算名次 = 1st & 4th（同队） + 2nd & 3rd（另一队）；赢家队牌级 +1。            


---

### P1 — 发牌（Deal Start）

|          |                                                                                     |
| -------- | ----------------------------------------------------------------------------------- |
| **进入条件** | 每个 Deal 开始                                                                          |
| **关键流程** | ① 洗牌（108 张）<br>② 若为 Match 第 1 Deal，切牌确定 **Starting Card** 并记录持有人<br>③ 顺时针发牌至每人 27 张 |
| **退出条件** | 四人 Hand = 27；首 Deal 额外记录 **Starting Card** 归属                                       |
| **下一阶段** | P2                                                                                  |

---

### P2 — 确定 Level & Trump

|          |                                                               |
| -------- | ------------------------------------------------------------- |
| **进入条件** | 发牌完毕                                                          |
| **关键流程** | ① 读取上一 Deal 胜方升后的 **Level**；若无则 2<br>② 将该点数的 8 张牌设为 **Trump** |
| **退出条件** | **Level** & Trump 写入 `Deal-ctx`                               |
| **下一阶段** | *若非首 Deal ➜ P3；否则 ➜ P4*                                       |

---

### P3 — **Tribute / Return Tribute / Tribute Immunity**

| 场景                                | **Tribute Immunity** 条件   | **Immunity** 失败时的流程                                                  |
| --------------------------------- | ------------------------- | -------------------------------------------------------------------- |
| **Double Down** (1 & 2 vs 3 & 4)  | 3 + 4 合计握两张 **Big Joker** | 3、4 各交 **Tribute Card** ➜ 1 先选、2 得余牌 ➜ 1、2 各 **Return Tribute** ≤ 10 |
| **Single Last** (1 & 3 vs 2 & 4)  | 4 单独握两张 **Big Joker**     | 4 交 **Tribute Card** ➜ 1 **Return Tribute** ≤ 10                     |
| **Partner Last** (1 & 4 vs 2 & 3) | 3 单独握两张 **Big Joker**     | 3 交 **Tribute Card** ➜ 1 **Return Tribute** ≤ 10                     |

*若满足 **Tribute Immunity** ➜ 跳过 Tribute/Return，直入 P4。*
上交**Tribute Card**时，需要给出自己手牌中，除了红桃trump外最大的牌
**Return Tribute**时，玩家可以任选一张自己的手牌（要求点数<=10）给出
上贡完毕后，每位玩家手里的牌数不变，仍为27张。
---


### P4 — 确定本 Deal 首出者

1. **首 Deal**：持有 **Starting Card** 者先出。
2. **非首 Deal**：按上一 Deal 结果 & 是否 Tribute Immunity，查表：

| 场景               | 无 Immunity      | 有 Immunity |
| ---------------- | --------------- | ---------- |
| **Double Down**  | 贡牌较大者；并列 ➜ 头游下家 | Seat 1     |
| **Single Last**  | Seat 4          | Seat 1     |
| **Partner Last** | Seat 3          | Seat 1     |

---

### P5 — 出牌环节（多 Trick 循环）

> **出牌环节的循环 ，参考以下伪代码的逻辑**

```pseudocode
WHILE true:                                   // —— Deal 主循环；每轮代表 1 个 Trick —— 
    //----------------------------------------------------------------
    // ① 领牌者出牌并初始化本 Trick 的状态
    //----------------------------------------------------------------
    tablePlay   ← playBy(leader)              // 领牌者必须先出一手合法牌；返回桌面牌型
    lastWinner  ← leader                      // 最近一次有效出牌者
    passSet     ← ∅                           // 本 Trick 已 PASS 的座次集合
    current     ← nextClockwise(leader)       // 顺时针找到下一个座次（可能已打空）

    //----------------------------------------------------------------
    // ② 进入单 Trick 循环，直到出现连续 PASS 或触发 Deal 结束
    //----------------------------------------------------------------
    WHILE true:                               // ——— 单 Trick 循环 ———
        // 2-1 跳过已出完牌的玩家
        WHILE current NOT IN activePlayers:   
            current ← nextClockwise(current)

        // 2-2 “当前玩家”决定是否跟牌
        IF canBeat(current, tablePlay):       // —— 跟牌分支 —— 
            tablePlay  ← playBy(current)      // 出更大的牌型并更新桌面
            lastWinner ← current              // 更新最近一次有效出牌者
            passSet    ← ∅                    // 任何人跟牌均清空 PASS 计数

            // 2-2-a 当前玩家打空 → 锁定名次
            IF handEmpty(current):            
                rankList.APPEND(current)      // 记录名次
                activePlayers.REMOVE(current) // 从活跃列表移除
                passSet.REMOVE(current)       // 修正②：保持 passSet 同步

                // -------- Deal 终止条件 #1：Double Down --------
                IF twoPlayersSameTeam(rankList):  
                    rankList.EXTEND(clockwiseRest(current)) // 依顺时针补齐 3-4 名
                    BREAK 2                 // 跳出 Trick 循环 + Deal 主循环
        ELSE:                               // —— PASS 分支 —— 
            passSet.ADD(current)            // 标记该玩家本 Trick 已 PASS

        // 2-3 判断 Trick 是否结束：除赢家外全部 PASS
        IF passSet.SIZE = activePlayers.SIZE – 1:
            // 确保新领牌者仍持有手牌
            leader ← nextClockwise(lastWinner)
            WHILE leader NOT IN activePlayers:
                leader ← nextClockwise(leader)
            BREAK                          // 跳出单 Trick 循环

        current ← nextClockwise(current)   // 继续轮到下一个座次
    //----------------------------------------------------------------
    // —— 单 Trick 结束；若 Deal 未终结则继续下一个 Trick —— 
    //----------------------------------------------------------------

    // -------- Deal 终止条件 #2：仅剩 1 名活跃玩家 --------
    IF activePlayers.SIZE = 1:              // 剩余者自动成为末游
        rankList.APPEND(activePlayers[0])   // 填入最后名次
        BREAK                               // 结束 Deal

    // Double Down 情况已通过 BREAK 2 跳出双重循环
END WHILE
```

`DEAL_END`：跳出全部循环 → P6。
*若任何时刻 `activePlayers.size == 1`，剩余者自动成末游，直接进入 P6。*

---

### P6 — 本 Deal 结算

| 名次组合             | 胜方升级     |
| ---------------- | -------- |
| **Double Down**  | +3 Level |
| **Single Last**  | +2 Level |
| **Partner Last** | +1 Level |

*更新胜方 **Level** ➜ 若 ≥ A，则 Match 结束；否则回到 P1 开启下一 Deal。*

---

### Match 终止条件

1. 任一队升至 **A-Level**；


---
                                 |
