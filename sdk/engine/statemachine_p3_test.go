package engine

import (
	"testing"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

func TestP3StateMachineDoubleDownFlow(t *testing.T) {
	t.Run("Complete Double Down state machine flow", func(t *testing.T) {
		// 创建测试环境
		eventBus := event.NewEventBus(100)
		players := []*domain.Player{
			domain.NewPlayer("p1", "Player1", domain.SeatEast),
			domain.NewPlayer("p2", "Player2", domain.SeatSouth),
			domain.NewPlayer("p3", "Player3", domain.SeatWest),
			domain.NewPlayer("p4", "Player4", domain.SeatNorth),
		}
		matchCtx := domain.NewMatchCtx("test-match", players, 12345)
		
		// 模拟第二局，设置上局排名为 Double Down 场景
		lastRankings := []domain.SeatID{domain.SeatEast, domain.SeatWest, domain.SeatSouth, domain.SeatNorth} // EastWest(1&2) vs SouthNorth(3&4)
		
		sm := NewDealStateMachine(matchCtx, eventBus)
		
		// 开始第二局
		err := sm.StartDeal(2, domain.SeatEast)
		if err != nil {
			t.Fatalf("Failed to start deal: %v", err)
		}
		
		// 发牌
		err = sm.DealCards()
		if err != nil {
			t.Fatalf("Failed to deal cards: %v", err)
		}
		
		// 确定Trump
		err = sm.DetermineTrump()
		if err != nil {
			t.Fatalf("Failed to determine trump: %v", err)
		}
		
		// 设置上局排名和大王数量（无免疫）
		sm.dealCtx = sm.dealCtx.WithLastRankings(lastRankings)
		playerBigJokers := map[domain.SeatID]int{
			domain.SeatSouth: 1, domain.SeatNorth: 0, // 败方队伍合计1张大王，无免疫
		}
		
		// 重新初始化贡牌系统
		sm.dealCtx = sm.dealCtx.InitializeTribute(playerBigJokers)
		
		// 验证场景识别
		if sm.dealCtx.TributeInfo.Scenario != domain.TributeScenarioDoubleDown {
			t.Errorf("Expected DoubleDown scenario, got %v", sm.dealCtx.TributeInfo.Scenario)
		}
		
		if sm.dealCtx.TributeInfo.HasImmunity {
			t.Error("Should not have immunity")
		}
		
		// 开始贡牌阶段
		err = sm.StartTribute()
		if err != nil {
			t.Fatalf("Failed to start tribute: %v", err)
		}
		
		if sm.currentPhase != PhaseTribute {
			t.Errorf("Expected Tribute phase, got %s", sm.currentPhase.String())
		}
		
		// 简化测试：直接设置贡牌信息，绕过实际的卡牌验证
		// 这是为了测试Double Down选择流程，而不是测试贡牌验证逻辑
		card3 := domain.NewCard(domain.Spades, domain.King)
		card4 := domain.NewCard(domain.Clubs, domain.Queen)
		
		// 直接设置已给出的贡牌
		sm.dealCtx.TributeInfo.GivenTributes[domain.SeatSouth] = card3
		sm.dealCtx.TributeInfo.GivenTributes[domain.SeatNorth] = card4
		sm.dealCtx.TributeInfo.Phase = domain.TributePhaseGiving
		
		// 手动进入选择阶段
		err = sm.StartTributeSelection()
		if err != nil {
			t.Fatalf("Failed to start tribute selection: %v", err)
		}
		
		// 验证进入选择阶段
		if sm.currentPhase != PhaseTributeSelection {
			t.Errorf("Expected TributeSelection phase, got %s", sm.currentPhase.String())
		}
		
		if sm.dealCtx.TributeInfo.Phase != domain.TributePhaseSelection {
			t.Errorf("Expected Selection phase, got %v", sm.dealCtx.TributeInfo.Phase)
		}
		
		// 验证可选卡牌
		if len(sm.dealCtx.TributeInfo.AvailableCards) != 2 {
			t.Errorf("Expected 2 available cards, got %d", len(sm.dealCtx.TributeInfo.AvailableCards))
		}
		
		// Player 1 选择来自 Player 3 的牌
		err = sm.SelectTributeCard(domain.SeatSouth)
		if err != nil {
			t.Fatalf("Failed to select tribute card: %v", err)
		}
		
		// 验证选择结果
		if sm.currentPhase != PhaseReturnTribute {
			t.Errorf("Expected ReturnTribute phase, got %s", sm.currentPhase.String())
		}
		
		if sm.dealCtx.TributeInfo.Phase != domain.TributePhaseReturning {
			t.Errorf("Expected Returning phase, got %v", sm.dealCtx.TributeInfo.Phase)
		}
		
		// 验证卡牌分配
		player1Hand := sm.matchCtx.GetPlayer(domain.SeatEast).GetHand()
		player2Hand := sm.matchCtx.GetPlayer(domain.SeatWest).GetHand()
		
		hasCard3 := false
		for _, card := range player1Hand {
			if card == card3 {
				hasCard3 = true
				break
			}
		}
		if !hasCard3 {
			t.Error("Player 1 should have received card from Player 3")
		}
		
		hasCard4 := false
		for _, card := range player2Hand {
			if card == card4 {
				hasCard4 = true
				break
			}
		}
		if !hasCard4 {
			t.Error("Player 2 should have received card from Player 4")
		}
		
		// 验证还贡要求
		expectedReturnRequests := map[domain.SeatID]domain.SeatID{
			domain.SeatEast: domain.SeatSouth, // Player 1 还给 Player 3
			domain.SeatWest: domain.SeatNorth, // Player 2 还给 Player 4
		}
		
		for from, to := range expectedReturnRequests {
			if sm.dealCtx.TributeInfo.ReturnRequests[from] != to {
				t.Errorf("Expected return %v->%v, got %v->%v", from, to, from, sm.dealCtx.TributeInfo.ReturnRequests[from])
			}
		}
		
		// 注意：在这个简化测试中，我们绕过了实际的卡牌转移过程
		// 所以不验证确切的卡牌数量，只验证逻辑流程正确性
		
		t.Log("P3 Double Down state machine flow completed successfully")
	})
}

func TestP3StateMachineImmunityFlow(t *testing.T) {
	t.Run("Complete immunity flow", func(t *testing.T) {
		// 创建测试环境
		eventBus := event.NewEventBus(100)
		players := []*domain.Player{
			domain.NewPlayer("p1", "Player1", domain.SeatEast),
			domain.NewPlayer("p2", "Player2", domain.SeatSouth),
			domain.NewPlayer("p3", "Player3", domain.SeatWest),
			domain.NewPlayer("p4", "Player4", domain.SeatNorth),
		}
		matchCtx := domain.NewMatchCtx("test-match", players, 12345)
		
		// 模拟Single Last场景且有免疫
		lastRankings := []domain.SeatID{domain.SeatEast, domain.SeatSouth, domain.SeatWest, domain.SeatNorth} // 1&3 vs 2&4
		
		sm := NewDealStateMachine(matchCtx, eventBus)
		
		// 开始第二局
		err := sm.StartDeal(2, domain.SeatEast)
		if err != nil {
			t.Fatalf("Failed to start deal: %v", err)
		}
		
		// 发牌
		err = sm.DealCards()
		if err != nil {
			t.Fatalf("Failed to deal cards: %v", err)
		}
		
		// 确定Trump
		err = sm.DetermineTrump()
		if err != nil {
			t.Fatalf("Failed to determine trump: %v", err)
		}
		
		// 设置上局排名
		sm.dealCtx = sm.dealCtx.WithLastRankings(lastRankings)
		
		// 给Player 4两张大王以获得免疫
		players[3].ClearHand()
		players[3].AddCards([]domain.Card{
			domain.NewJoker(domain.BigJoker),
			domain.NewJoker(domain.BigJoker),
		})
		// 发牌给其他玩家确保有27张牌
		for i := 0; i < 25; i++ {
			players[3].AddCards([]domain.Card{domain.NewCard(domain.Hearts, domain.Two)})
		}
		
		playerBigJokers := map[domain.SeatID]int{
			domain.SeatNorth: 2, // 最后一名(4)单独握2张大王，有免疫
		}
		
		// 重新初始化贡牌系统
		sm.dealCtx = sm.dealCtx.InitializeTribute(playerBigJokers)
		
		// 验证场景识别
		if sm.dealCtx.TributeInfo.Scenario != domain.TributeScenarioSingleLast {
			t.Errorf("Expected SingleLast scenario, got %v", sm.dealCtx.TributeInfo.Scenario)
		}
		
		if !sm.dealCtx.TributeInfo.HasImmunity {
			t.Error("Should have immunity")
		}
		
		// 开始贡牌阶段（应该跳过）
		err = sm.StartTribute()
		if err != nil {
			t.Fatalf("Failed to start tribute: %v", err)
		}
		
		// 验证直接跳到FirstPlay阶段
		if sm.currentPhase != PhaseFirstPlay {
			t.Errorf("Expected FirstPlay phase due to immunity, got %s", sm.currentPhase.String())
		}
		
		if sm.dealCtx.TributeInfo.Phase != domain.TributePhaseCompleted {
			t.Errorf("Expected Completed phase, got %v", sm.dealCtx.TributeInfo.Phase)
		}
		
		t.Log("P3 immunity flow completed successfully")
	})
}