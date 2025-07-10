package domain

import (
	"fmt"
	"sort"
)

// TributeScenario 表示贡牌场景
type TributeScenario int

const (
	TributeScenarioNone TributeScenario = iota
	TributeScenarioDoubleDown  // 1&2 vs 3&4 (连庄)
	TributeScenarioSingleLast  // 1&3 vs 2&4 (单落)
	TributeScenarioPartnerLast // 1&4 vs 2&3 (对落)
)

func (ts TributeScenario) String() string {
	switch ts {
	case TributeScenarioNone:
		return "None"
	case TributeScenarioDoubleDown:
		return "DoubleDown"
	case TributeScenarioSingleLast:
		return "SingleLast"
	case TributeScenarioPartnerLast:
		return "PartnerLast"
	default:
		return "Unknown"
	}
}

// TributePhase 表示贡牌阶段
type TributePhase int

const (
	TributePhaseIdle TributePhase = iota
	TributePhaseRequested
	TributePhaseGiving
	TributePhaseSelection  // Double Down 卡牌选择阶段
	TributePhaseReturning
	TributePhaseCompleted
)

func (tp TributePhase) String() string {
	switch tp {
	case TributePhaseIdle:
		return "Idle"
	case TributePhaseRequested:
		return "Requested"
	case TributePhaseGiving:
		return "Giving"
	case TributePhaseSelection:
		return "Selection"
	case TributePhaseReturning:
		return "Returning"
	case TributePhaseCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}

// TributeInfo 存储贡牌信息
type TributeInfo struct {
	Scenario          TributeScenario
	HasImmunity       bool
	TributeRequests   map[SeatID]SeatID // from -> to
	ReturnRequests    map[SeatID]SeatID // from -> to
	GivenTributes     map[SeatID]Card   // from -> card
	ReturnedTributes  map[SeatID]Card   // from -> card
	Phase             TributePhase
	
	// Double Down 特殊字段
	AvailableCards    map[SeatID]Card   // 可选择的贡牌 (giver -> card) for Double Down
	SelectedCards     map[SeatID]Card   // 已选择的贡牌 (receiver -> card) for Double Down
	ActualReceivers   map[SeatID]SeatID // 实际接收者映射 (giver -> actual_receiver) for return tribute
}

// NewTributeInfo 创建新的贡牌信息
func NewTributeInfo(scenario TributeScenario, hasImmunity bool) *TributeInfo {
	return &TributeInfo{
		Scenario:         scenario,
		HasImmunity:      hasImmunity,
		TributeRequests:  make(map[SeatID]SeatID),
		ReturnRequests:   make(map[SeatID]SeatID),
		GivenTributes:    make(map[SeatID]Card),
		ReturnedTributes: make(map[SeatID]Card),
		Phase:            TributePhaseIdle,
		AvailableCards:   make(map[SeatID]Card),
		SelectedCards:    make(map[SeatID]Card),
		ActualReceivers:  make(map[SeatID]SeatID),
	}
}

// DetermineTributeScenario 根据上局排名确定贡牌场景
func DetermineTributeScenario(lastRankings []SeatID) TributeScenario {
	if len(lastRankings) != 4 {
		return TributeScenarioNone // 首局或数据不完整
	}

	first := lastRankings[0]
	second := lastRankings[1]
	third := lastRankings[2]
	fourth := lastRankings[3]

	// 检查第一和第二名是否为同队
	if GetTeamFromSeat(first) == GetTeamFromSeat(second) {
		return TributeScenarioDoubleDown // 连庄：1&2 vs 3&4
	}

	// 检查第三和第四名是否为同队（被升级队连败）
	if GetTeamFromSeat(third) == GetTeamFromSeat(fourth) {
		// 需要进一步判断是单落还是对落
		// 检查第一和第三名是否为同队
		if GetTeamFromSeat(first) == GetTeamFromSeat(third) {
			return TributeScenarioSingleLast // 单落：1&3 vs 2&4
		} else {
			return TributeScenarioPartnerLast // 对落：1&4 vs 2&3
		}
	}

	// 其他情况（如1&3获胜或1&4获胜）
	if GetTeamFromSeat(first) == GetTeamFromSeat(third) {
		return TributeScenarioSingleLast // 单落：1&3 vs 2&4
	}

	if GetTeamFromSeat(first) == GetTeamFromSeat(fourth) {
		return TributeScenarioPartnerLast // 对落：1&4 vs 2&3
	}

	return TributeScenarioNone // 不应该发生的情况
}

// CheckTributeImmunity 检查是否有贡牌免疫
func CheckTributeImmunity(scenario TributeScenario, playerBigJokers map[SeatID]int, lastRankings []SeatID) bool {
	if len(lastRankings) != 4 {
		return false
	}
	
	switch scenario {
	case TributeScenarioDoubleDown:
		// 败方队伍(3+4)合计握两张大王
		third := lastRankings[2]
		fourth := lastRankings[3]
		return playerBigJokers[third]+playerBigJokers[fourth] >= 2
	case TributeScenarioSingleLast:
		// 最后一名(4)单独握两张大王
		fourth := lastRankings[3]
		return playerBigJokers[fourth] >= 2
	case TributeScenarioPartnerLast:
		// 第三名(3)单独握两张大王
		third := lastRankings[2]
		return playerBigJokers[third] >= 2
	default:
		return false
	}
}

// CalculateTributeRequirements 计算贡牌要求
func CalculateTributeRequirements(scenario TributeScenario, lastRankings []SeatID) map[SeatID]SeatID {
	requirements := make(map[SeatID]SeatID)
	
	if len(lastRankings) != 4 {
		return requirements
	}
	
	first := lastRankings[0]
	second := lastRankings[1]
	third := lastRankings[2]
	fourth := lastRankings[3]

	switch scenario {
	case TributeScenarioDoubleDown:
		// 败方队伍(3、4)各交给胜方队伍(1、2)
		requirements[third] = first   // 3->1
		requirements[fourth] = second // 4->2
	case TributeScenarioSingleLast:
		// 最后一名(4)交给第一名(1)
		requirements[fourth] = first // 4->1
	case TributeScenarioPartnerLast:
		// 第三名(3)交给第一名(1)
		requirements[third] = first // 3->1
	}

	return requirements
}

// CalculateReturnRequirements 计算还贡要求
func CalculateReturnRequirements(scenario TributeScenario, lastRankings []SeatID) map[SeatID]SeatID {
	requirements := make(map[SeatID]SeatID)
	
	if len(lastRankings) != 4 {
		return requirements
	}
	
	first := lastRankings[0]
	second := lastRankings[1]
	third := lastRankings[2]
	fourth := lastRankings[3]

	switch scenario {
	case TributeScenarioDoubleDown:
		// 胜方队伍(1、2)各还给败方队伍(3、4)
		// Note: For Double Down, return assignments will be determined after card selection
		requirements[first] = third   // 1->3 (initial assignment)
		requirements[second] = fourth // 2->4 (initial assignment)
	case TributeScenarioSingleLast:
		// 第一名(1)还给最后一名(4)
		requirements[first] = fourth // 1->4
	case TributeScenarioPartnerLast:
		// 第一名(1)还给第三名(3)
		requirements[first] = third // 1->3
	}

	return requirements
}

// SelectTributeCard 选择贡牌（除了红桃trump外最大的牌）
func SelectTributeCard(hand []Card, trump Rank) (Card, bool) {
	if len(hand) == 0 {
		return Card{}, false
	}

	// 过滤掉红桃trump（但保留其他花色的trump）
	validCards := make([]Card, 0)
	for _, card := range hand {
		if !(card.Suit == Hearts && card.Rank == trump) {
			validCards = append(validCards, card)
		}
	}

	if len(validCards) == 0 {
		// 只有红桃trump，无法合法选择
		return hand[0], false
	}

	// 按照牌值排序，选择最大的
	// 需要特别处理trump牌（非红桃的trump也应该排在前面）
	sort.Slice(validCards, func(i, j int) bool {
		valueI := getTributeCardValue(validCards[i], trump)
		valueJ := getTributeCardValue(validCards[j], trump)
		return valueI > valueJ
	})

	return validCards[0], true
}

// getTributeCardValue 获取贡牌场景下的牌值（trump牌优先级最高）
func getTributeCardValue(card Card, trump Rank) int {
	if card.IsJoker() {
		if card.Rank == SmallJoker {
			return 1000
		}
		if card.Rank == BigJoker {
			return 1001
		}
	}
	
	// trump牌（非红桃）有最高优先级，按花色排序
	if card.Rank == trump && card.Suit != Hearts {
		return 500 + int(card.Suit) // 黑桃(3) > 梅花(2) > 方片(1)
	}
	
	return int(card.Rank)
}

// IsValidReturnTributeCard 验证还贡牌是否有效（点数<=10）
func IsValidReturnTributeCard(hand []Card, selectedCard Card) bool {
	// 检查牌是否在手中
	hasCard := false
	for _, card := range hand {
		if card == selectedCard {
			hasCard = true
			break
		}
	}
	if !hasCard {
		return false
	}

	// 检查点数是否<=10 (注意：J=11, Q=12, K=13, A=14)
	// Ten=10是有效的
	if selectedCard.Rank == Jack || selectedCard.Rank == Queen || selectedCard.Rank == King || selectedCard.Rank == Ace {
		return false
	}
	
	// 王牌不能用于还贡
	if selectedCard.IsJoker() {
		return false
	}

	return true
}

// CountBigJokers 计算手中大王数量
func CountBigJokers(hand []Card) int {
	count := 0
	for _, card := range hand {
		if card.IsJoker() && card.Rank == BigJoker {
			count++
		}
	}
	return count
}

// ValidateTributeCard 验证贡牌是否符合规则
func ValidateTributeCard(hand []Card, selectedCard Card, trump Rank) error {
	// 检查牌是否在手中
	hasCard := false
	for _, card := range hand {
		if card == selectedCard {
			hasCard = true
			break
		}
	}
	if !hasCard {
		return fmt.Errorf("selected card not in hand")
	}

	// 检查是否是红桃trump
	if selectedCard.Suit == Hearts && selectedCard.Rank == trump {
		return fmt.Errorf("cannot tribute Hearts trump card")
	}

	// 检查是否为最大牌
	expectedCard, valid := SelectTributeCard(hand, trump)
	if !valid {
		return fmt.Errorf("no valid tribute card available")
	}

	if selectedCard != expectedCard {
		return fmt.Errorf("must tribute the highest card (expected %v, got %v)", expectedCard, selectedCard)
	}

	return nil
}

// GetTributeCardCandidates 获取可能的贡牌候选（调试用）
func GetTributeCardCandidates(hand []Card, trump Rank) []Card {
	candidates := make([]Card, 0)
	for _, card := range hand {
		if !(card.Suit == Hearts && card.Rank == trump) {
			candidates = append(candidates, card)
		}
	}

	// 按照牌值排序
	sort.Slice(candidates, func(i, j int) bool {
		return getCardValue(candidates[i], trump) > getCardValue(candidates[j], trump)
	})

	return candidates
}

// GetReturnTributeCardCandidates 获取可能的还贡牌候选
func GetReturnTributeCardCandidates(hand []Card) []Card {
	candidates := make([]Card, 0)
	for _, card := range hand {
		if int(card.Rank) <= 10 {
			candidates = append(candidates, card)
		}
	}

	// 按照牌值排序（从小到大，还贡通常选择较小的牌）
	sort.Slice(candidates, func(i, j int) bool {
		return int(candidates[i].Rank) < int(candidates[j].Rank)
	})

	return candidates
}

// IsTributeComplete 检查贡牌是否完成
func (ti *TributeInfo) IsTributeComplete() bool {
	if ti.HasImmunity {
		return true
	}

	// 检查所有要求的贡牌是否都已给出
	for from := range ti.TributeRequests {
		if _, exists := ti.GivenTributes[from]; !exists {
			return false
		}
	}

	return true
}

// IsReturnComplete 检查还贡是否完成
func (ti *TributeInfo) IsReturnComplete() bool {
	if ti.HasImmunity {
		return true
	}

	// 检查所有要求的还贡是否都已给出
	for from := range ti.ReturnRequests {
		if _, exists := ti.ReturnedTributes[from]; !exists {
			return false
		}
	}

	return true
}

// GetTributeOrder 获取贡牌顺序（Double Down场景中1先选，2得余牌）
func GetTributeOrder(scenario TributeScenario) []SeatID {
	switch scenario {
	case TributeScenarioDoubleDown:
		return []SeatID{SeatEast, SeatSouth} // 1先选，2得余牌
	case TributeScenarioSingleLast:
		return []SeatID{SeatEast} // 只有1选择
	case TributeScenarioPartnerLast:
		return []SeatID{SeatEast} // 只有1选择
	default:
		return []SeatID{}
	}
}

// GetAvailableTributeCards 获取可用的贡牌选择
func GetAvailableTributeCards(givenTributes map[SeatID]Card, scenario TributeScenario) []Card {
	cards := make([]Card, 0)
	for _, card := range givenTributes {
		cards = append(cards, card)
	}
	return cards
}

// PrepareDoubleDownSelection 准备Double Down场景的卡牌选择
func (ti *TributeInfo) PrepareDoubleDownSelection(lastRankings []SeatID) {
	if ti.Scenario != TributeScenarioDoubleDown || len(lastRankings) != 4 {
		return
	}
	
	third := lastRankings[2]
	fourth := lastRankings[3]
	
	// 将已给出的贡牌移到可选择列表
	if card, exists := ti.GivenTributes[third]; exists {
		ti.AvailableCards[third] = card
	}
	if card, exists := ti.GivenTributes[fourth]; exists {
		ti.AvailableCards[fourth] = card
	}
	
	ti.Phase = TributePhaseSelection
}

// SelectTributeCardForDoubleDown Player 1在Double Down场景中选择贡牌
func (ti *TributeInfo) SelectTributeCardForDoubleDown(giver SeatID, lastRankings []SeatID) error {
	if ti.Scenario != TributeScenarioDoubleDown || len(lastRankings) != 4 {
		return fmt.Errorf("not a Double Down scenario")
	}
	
	if ti.Phase != TributePhaseSelection {
		return fmt.Errorf("not in selection phase")
	}
	
	first := lastRankings[0]
	second := lastRankings[1]
	third := lastRankings[2]
	fourth := lastRankings[3]
	
	// 检查是否是有效的贡牌者
	if giver != third && giver != fourth {
		return fmt.Errorf("invalid giver for selection")
	}
	
	// 检查卡牌是否可选
	selectedCard, exists := ti.AvailableCards[giver]
	if !exists {
		return fmt.Errorf("no card available from giver %s", giver.String())
	}
	
	// Player 1 选择这张牌
	ti.SelectedCards[first] = selectedCard
	ti.ActualReceivers[giver] = first
	
	// Player 2 自动获得剩余的牌
	var remainingGiver SeatID
	if giver == third {
		remainingGiver = fourth
	} else {
		remainingGiver = third
	}
	
	if card, exists := ti.AvailableCards[remainingGiver]; exists {
		ti.SelectedCards[second] = card
		ti.ActualReceivers[remainingGiver] = second
	}
	
	// 清空可选卡牌列表
	ti.AvailableCards = make(map[SeatID]Card)
	
	// 更新还贡要求基于实际分配
	ti.ReturnRequests = make(map[SeatID]SeatID)
	// 根据实际接收情况设置还贡要求
	for giver, receiver := range ti.ActualReceivers {
		ti.ReturnRequests[receiver] = giver // receiver 还给 giver
	}
	
	// 转到还贡阶段
	ti.Phase = TributePhaseReturning
	
	return nil
}

// IsSelectionRequired 检查是否需要进行卡牌选择
func (ti *TributeInfo) IsSelectionRequired() bool {
	return ti.Scenario == TributeScenarioDoubleDown && !ti.HasImmunity && ti.Phase == TributePhaseSelection
}

// IsSelectionComplete 检查选择是否完成
func (ti *TributeInfo) IsSelectionComplete() bool {
	if ti.Scenario != TributeScenarioDoubleDown {
		return true // 非Double Down场景不需要选择
	}
	return len(ti.SelectedCards) == 2 // Double Down需要选择2张卡
}