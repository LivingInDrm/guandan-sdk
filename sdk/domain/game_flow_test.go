package domain

import (
	"testing"
	"time"
)

// 游戏流程相关的测试

func TestTributeSystemRules(t *testing.T) {
	testCases := []struct {
		name           string
		dealNumber     int
		lastRankings   []SeatID // 上局排名
		expectTribute  bool
		tributeFrom    SeatID
		tributeTo      SeatID
	}{
		{
			name:           "First deal - no tribute required",
			dealNumber:     1,
			lastRankings:   nil,
			expectTribute:  false,
		},
		{
			name:           "Second deal - tribute based on rankings",
			dealNumber:     2,
			lastRankings:   []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}, // East wins, North last
			expectTribute:  true,
			tributeFrom:    SeatNorth, // Last place
			tributeTo:      SeatEast,  // First place
		},
		{
			name:           "Normal deal progression",
			dealNumber:     3,
			lastRankings:   []SeatID{SeatWest, SeatNorth, SeatEast, SeatSouth},
			expectTribute:  true,
			tributeFrom:    SeatSouth,
			tributeTo:      SeatWest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dealCtx := NewDealCtx(tc.dealNumber, Two, SeatEast)
			
			// 验证是否需要贡牌
			shouldTribute := !dealCtx.IsFirstDeal
			if shouldTribute != tc.expectTribute {
				t.Errorf("Expected tribute requirement %v, got %v for deal %d", 
					tc.expectTribute, shouldTribute, tc.dealNumber)
			}

			// 如果需要贡牌，测试贡牌逻辑
			if tc.expectTribute {
				// 这里可以添加具体的贡牌规则测试
				// 当前domain层可能没有完整的贡牌逻辑实现
				t.Logf("Deal %d should require tribute from %v to %v", 
					tc.dealNumber, tc.tributeFrom, tc.tributeTo)
			}
		})
	}
}

func TestRankingSystem(t *testing.T) {
	dealCtx := NewDealCtx(1, Two, SeatEast)

	// 测试添加到排名列表
	dealCtx = dealCtx.AddToRankList(SeatNorth) // 第一名
	dealCtx = dealCtx.AddToRankList(SeatEast)  // 第二名  
	dealCtx = dealCtx.AddToRankList(SeatSouth) // 第三名
	// SeatWest 第四名（未完成）

	// 验证排名位置
	if pos := dealCtx.GetRankPosition(SeatNorth); pos != 1 {
		t.Errorf("Expected North in position 1, got %d", pos)
	}
	if pos := dealCtx.GetRankPosition(SeatEast); pos != 2 {
		t.Errorf("Expected East in position 2, got %d", pos)
	}
	if pos := dealCtx.GetRankPosition(SeatSouth); pos != 3 {
		t.Errorf("Expected South in position 3, got %d", pos)
	}
	if pos := dealCtx.GetRankPosition(SeatWest); pos != 0 {
		t.Errorf("Expected West not ranked (0), got %d", pos)
	}

	// 验证排名列表长度
	if len(dealCtx.RankList) != 3 {
		t.Errorf("Expected 3 players in rank list, got %d", len(dealCtx.RankList))
	}
}

func TestGameStateTransitions(t *testing.T) {
	// 测试比赛状态转换
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)

	// 初始状态应该是Created
	if matchCtx.State != MatchStateCreated {
		t.Errorf("Expected initial state Created, got %v", matchCtx.State)
	}

	// 转换到InProgress
	matchCtx = matchCtx.WithState(MatchStateInProgress)
	if matchCtx.State != MatchStateInProgress {
		t.Errorf("Expected state InProgress after transition, got %v", matchCtx.State)
	}

	// 设置获胜者并结束比赛
	matchCtx = matchCtx.WithWinner(TeamEastWest)
	if matchCtx.State != MatchStateFinished {
		t.Errorf("Expected state Finished after setting winner, got %v", matchCtx.State)
	}
	if !matchCtx.IsFinished() {
		t.Error("Match should be finished after setting winner")
	}
	if matchCtx.EndTime == nil {
		t.Error("End time should be set when match finishes")
	}
}

func TestDealStateProgression(t *testing.T) {
	dealCtx := NewDealCtx(1, Three, SeatSouth)

	// 验证状态顺序
	states := []DealState{
		DealStateCreated,
		DealStateDealt,
		DealStateTribute,
		DealStateFirstPlay,
		DealStateInProgress,
		DealStateFinished,
	}

	currentCtx := dealCtx
	for i, state := range states {
		if i == 0 {
			// 初始状态验证
			if currentCtx.State != state {
				t.Errorf("Expected initial state %v, got %v", state, currentCtx.State)
			}
		} else {
			// 状态转换
			currentCtx = currentCtx.WithState(state)
			if currentCtx.State != state {
				t.Errorf("Expected state %v after transition, got %v", state, currentCtx.State)
			}
		}
	}

	// 验证最终状态
	if !currentCtx.IsFinished() {
		t.Error("Deal should be finished in final state")
	}
}

func TestTrickProgression(t *testing.T) {
	trickCtx := NewTrickCtx(1, SeatEast)

	// 初始状态验证
	if trickCtx.State != TrickStateActive {
		t.Errorf("Expected initial trick state Active, got %v", trickCtx.State)
	}
	if trickCtx.GetActivePlayerCount() != 4 {
		t.Errorf("Expected 4 active players initially, got %d", trickCtx.GetActivePlayerCount())
	}

	// 玩家依次过牌
	trickCtx = trickCtx.WithPlayerPassed(SeatSouth)
	if trickCtx.GetActivePlayerCount() != 3 {
		t.Errorf("Expected 3 active players after one pass, got %d", trickCtx.GetActivePlayerCount())
	}

	trickCtx = trickCtx.WithPlayerPassed(SeatWest)
	if trickCtx.GetActivePlayerCount() != 2 {
		t.Errorf("Expected 2 active players after two passes, got %d", trickCtx.GetActivePlayerCount())
	}

	trickCtx = trickCtx.WithPlayerPassed(SeatNorth)
	if trickCtx.GetActivePlayerCount() != 1 {
		t.Errorf("Expected 1 active player after three passes, got %d", trickCtx.GetActivePlayerCount())
	}

	// 应该可以结束了
	if !trickCtx.ShouldFinish() {
		t.Error("Trick should be ready to finish with only 1 active player")
	}

	// 设置获胜者
	trickCtx = trickCtx.WithWinner(SeatEast)
	if !trickCtx.IsFinished() {
		t.Error("Trick should be finished after setting winner")
	}
}

func TestPlayHistoryTracking(t *testing.T) {
	trickCtx := NewTrickCtx(1, SeatEast)

	// 创建一些出牌记录
	play1 := TrickPlay{
		Player:    SeatEast,
		Cards:     []Card{NewCard(Hearts, Ace)},
		CardGroup: NewCardGroup([]Card{NewCard(Hearts, Ace)}),
		Timestamp: time.Now(),
	}

	play2 := TrickPlay{
		Player:    SeatSouth,
		Cards:     []Card{NewCard(Spades, King)},
		CardGroup: NewCardGroup([]Card{NewCard(Spades, King)}),
		Timestamp: time.Now(),
	}

	// 添加出牌历史
	trickCtx = trickCtx.WithPlayHistory(play1)
	if len(trickCtx.PlayHistory) != 1 {
		t.Errorf("Expected 1 play in history, got %d", len(trickCtx.PlayHistory))
	}

	trickCtx = trickCtx.WithPlayHistory(play2)
	if len(trickCtx.PlayHistory) != 2 {
		t.Errorf("Expected 2 plays in history, got %d", len(trickCtx.PlayHistory))
	}

	// 验证历史记录顺序
	if trickCtx.PlayHistory[0].Player != SeatEast {
		t.Errorf("Expected first play from East, got %v", trickCtx.PlayHistory[0].Player)
	}
	if trickCtx.PlayHistory[1].Player != SeatSouth {
		t.Errorf("Expected second play from South, got %v", trickCtx.PlayHistory[1].Player)
	}
}

func TestPlayerTurnRotation(t *testing.T) {
	// 测试玩家轮次系统
	testCases := []struct {
		current SeatID
		next    SeatID
	}{
		{SeatEast, SeatSouth},
		{SeatSouth, SeatWest},
		{SeatWest, SeatNorth},
		{SeatNorth, SeatEast}, // 循环回到East
	}

	for _, tc := range testCases {
		t.Run(tc.current.String(), func(t *testing.T) {
			next := tc.current.Next()
			if next != tc.next {
				t.Errorf("Expected %v.Next() = %v, got %v", tc.current, tc.next, next)
			}
		})
	}
}

func TestTeamVictoryConditions(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("victory-test", players, 54321)

	// 模拟东西队获胜
	eastWestTeam := matchCtx.GetTeam(TeamEastWest)
	southNorthTeam := matchCtx.GetTeam(TeamSouthNorth)

	if eastWestTeam == nil || southNorthTeam == nil {
		t.Fatal("Teams should be properly initialized")
	}

	// 验证队伍成员
	eastWestPlayers := eastWestTeam.GetPlayers()
	southNorthPlayers := southNorthTeam.GetPlayers()

	if len(eastWestPlayers) != 2 {
		t.Errorf("East-West team should have 2 players, got %d", len(eastWestPlayers))
	}
	if len(southNorthPlayers) != 2 {
		t.Errorf("South-North team should have 2 players, got %d", len(southNorthPlayers))
	}

	// 设置获胜队伍
	matchCtx = matchCtx.WithWinner(TeamEastWest)
	if matchCtx.Winner == nil || *matchCtx.Winner != TeamEastWest {
		t.Error("East-West team should be set as winner")
	}
}

func TestLevelProgression(t *testing.T) {
	// 测试等级系统（如果已实现）
	// 当前domain层可能只有基础的等级概念，高级升级规则可能在engine层
	
	dealCtx := NewDealCtx(1, Two, SeatEast)
	
	// 验证当前等级就是主牌
	if dealCtx.CurrentLevel != dealCtx.Trump {
		t.Errorf("Current level should match trump, got level=%v, trump=%v", 
			dealCtx.CurrentLevel, dealCtx.Trump)
	}

	// 不同等级的发牌测试
	levels := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}
	
	for i, level := range levels {
		dealCtx := NewDealCtx(i+1, level, SeatEast)
		if dealCtx.CurrentLevel != level {
			t.Errorf("Deal %d should have level %v, got %v", i+1, level, dealCtx.CurrentLevel)
		}
		if dealCtx.Trump != level {
			t.Errorf("Deal %d should have trump %v, got %v", i+1, level, dealCtx.Trump)
		}
	}
}

// 性能测试
func BenchmarkStateTransitions(b *testing.B) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("bench-test", players, 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟状态转换
		_ = matchCtx.WithState(MatchStateInProgress)
		_ = matchCtx.WithCurrentDeal(i)
		_ = matchCtx.WithWinner(TeamEastWest)
	}
}

func BenchmarkTrickProgression(b *testing.B) {
	trickCtx := NewTrickCtx(1, SeatEast)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟墩的进行
		ctx := trickCtx.WithCurrentPlayer(SeatSouth)
		ctx = ctx.WithPlayerPassed(SeatWest)
		ctx = ctx.WithWinner(SeatEast)
	}
}