package domain

import (
	"testing"
)

func TestDetermineFirstPlayer_FirstDeal(t *testing.T) {
	// 创建测试用的MatchCtx和DealCtx
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 首Deal测试
	dealCtx := NewDealCtx(1, Two, SeatEast) // 临时首出者
	
	startingCardHolder := SeatWest
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, startingCardHolder)
	
	if firstPlayer != startingCardHolder {
		t.Errorf("首Deal应该由Starting Card持有者先出，期望: %v, 实际: %v", startingCardHolder, firstPlayer)
	}
}

func TestDetermineFirstPlayer_DoubleDownWithoutImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, West第2, South第3, North第4 (Double Down场景: 1&2同队 vs 3&4同队)
	lastRankings := []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 无免疫
	tributeInfo := NewTributeInfo(TributeScenarioDoubleDown, false)
	// 模拟贡牌：South贡给East，North贡给West
	tributeInfo.GivenTributes[SeatSouth] = Card{Suit: Spades, Rank: King}  // 较大的牌
	tributeInfo.GivenTributes[SeatNorth] = Card{Suit: Hearts, Rank: Queen} // 较小的牌
	
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// South贡的牌更大，应该选择接收South贡牌的人(East)
	expectedFirstPlayer := SeatEast
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("DoubleDown无免疫场景下应该选择贡牌较大者的接收者，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_DoubleDownWithImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, West第2, South第3, North第4 (Double Down场景: 1&2同队 vs 3&4同队)
	lastRankings := []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 有免疫
	tributeInfo := NewTributeInfo(TributeScenarioDoubleDown, true)
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 有免疫，应该选择上轮第1名
	expectedFirstPlayer := SeatEast
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("DoubleDown有免疫场景下应该选择上轮第1名，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_SingleLastWithoutImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, South第2, West第3, North第4 (Single Last场景: 1&3同队 vs 2&4同队)
	lastRankings := []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 无免疫
	tributeInfo := NewTributeInfo(TributeScenarioSingleLast, false)
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 无免疫，应该选择上轮第4名
	expectedFirstPlayer := SeatNorth
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("SingleLast无免疫场景下应该选择上轮第4名，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_SingleLastWithImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, South第2, West第3, North第4 (Single Last场景: 1&3同队 vs 2&4同队)
	lastRankings := []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 有免疫
	tributeInfo := NewTributeInfo(TributeScenarioSingleLast, true)
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 有免疫，应该选择上轮第1名
	expectedFirstPlayer := SeatEast
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("SingleLast有免疫场景下应该选择上轮第1名，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_PartnerLastWithoutImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, South第2, North第3, West第4 (Partner Last场景: 1&4同队 vs 2&3同队)
	lastRankings := []SeatID{SeatEast, SeatSouth, SeatNorth, SeatWest}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 无免疫
	tributeInfo := NewTributeInfo(TributeScenarioPartnerLast, false)
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 无免疫，应该选择上轮第3名
	expectedFirstPlayer := SeatNorth
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("PartnerLast无免疫场景下应该选择上轮第3名，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_PartnerLastWithImmunity(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, South第2, North第3, West第4 (Partner Last场景: 1&4同队 vs 2&3同队)
	lastRankings := []SeatID{SeatEast, SeatSouth, SeatNorth, SeatWest}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 有免疫
	tributeInfo := NewTributeInfo(TributeScenarioPartnerLast, true)
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 有免疫，应该选择上轮第1名
	expectedFirstPlayer := SeatEast
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("PartnerLast有免疫场景下应该选择上轮第1名，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_DoubleDownTieBreaker(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 上一Deal排名：East第1, West第2, South第3, North第4 (Double Down场景: 1&2同队 vs 3&4同队)
	lastRankings := []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth}
	
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(lastRankings)
	
	// 创建贡牌信息 - 无免疫，两张贡牌一样大
	tributeInfo := NewTributeInfo(TributeScenarioDoubleDown, false)
	// 模拟贡牌：South和North都贡Queen，一样大
	tributeInfo.GivenTributes[SeatSouth] = Card{Suit: Spades, Rank: Queen}
	tributeInfo.GivenTributes[SeatNorth] = Card{Suit: Hearts, Rank: Queen}
	
	dealCtx = dealCtx.WithTributeInfo(tributeInfo)
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 两张贡牌一样大，应该选择第1名的顺时针下家(South)
	expectedFirstPlayer := SeatSouth
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("DoubleDown贡牌一样大时应该选择第1名的顺时针下家，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestDetermineFirstPlayer_MissingData(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 测试数据不完整的情况
	dealCtx := NewDealCtx(2, Three, SeatEast) // 临时首出者
	dealCtx = dealCtx.WithLastRankings(nil) // 没有上局排名
	
	firstPlayer := DetermineFirstPlayer(matchCtx, dealCtx, SeatEast)
	
	// 数据不完整时应该返回默认值East
	expectedFirstPlayer := SeatEast
	if firstPlayer != expectedFirstPlayer {
		t.Errorf("数据不完整时应该返回默认值，期望: %v, 实际: %v", expectedFirstPlayer, firstPlayer)
	}
}

func TestNewDealCtxWithAutoFirstPlayer(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// 测试首Deal自动确定首出者
	startingCardHolder := SeatWest
	dealCtx := NewDealCtxWithAutoFirstPlayer(matchCtx, 1, Two, nil, startingCardHolder)
	
	if dealCtx.FirstPlayer != startingCardHolder {
		t.Errorf("NewDealCtxWithAutoFirstPlayer应该正确设置首Deal的首出者，期望: %v, 实际: %v", 
			startingCardHolder, dealCtx.FirstPlayer)
	}
	
	if !dealCtx.IsFirstDeal {
		t.Error("NewDealCtxWithAutoFirstPlayer应该正确设置IsFirstDeal标志")
	}
	
	// 测试非首Deal自动确定首出者
	lastRankings := []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}
	dealCtx2 := NewDealCtxWithAutoFirstPlayer(matchCtx, 2, Three, lastRankings, SeatEast)
	
	if dealCtx2.IsFirstDeal {
		t.Error("第2个Deal不应该是FirstDeal")
	}
	
	if len(dealCtx2.LastRankings) != 4 {
		t.Errorf("LastRankings应该被正确设置，期望长度: 4, 实际长度: %d", len(dealCtx2.LastRankings))
	}
}