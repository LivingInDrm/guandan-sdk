package engine

import (
	"fmt"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

type DealPhase int

const (
	PhaseIdle DealPhase = iota
	PhaseCreated
	PhaseCardsDealt
	PhaseTrumpDecision
	PhaseTribute
	PhaseTributeSelection  // Double Down 卡牌选择阶段
	PhaseReturnTribute
	PhaseFirstPlay
	PhaseInProgress
	PhaseRankList
	PhaseFinished
)

func (p DealPhase) String() string {
	switch p {
	case PhaseIdle:
		return "Idle"
	case PhaseCreated:
		return "Created"
	case PhaseCardsDealt:
		return "CardsDealt"
	case PhaseTrumpDecision:
		return "TrumpDecision"
	case PhaseTribute:
		return "Tribute"
	case PhaseTributeSelection:
		return "TributeSelection"
	case PhaseReturnTribute:
		return "ReturnTribute"
	case PhaseFirstPlay:
		return "FirstPlay"
	case PhaseInProgress:
		return "InProgress"
	case PhaseRankList:
		return "RankList"
	case PhaseFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type DealStateMachine struct {
	currentPhase DealPhase
	matchCtx     *domain.MatchCtx
	dealCtx      *domain.DealCtx
	trickCtx     *domain.TrickCtx
	eventBus     *event.EventBus
	deck         *domain.Deck
	startingCard *domain.Card
	startingCardHolder domain.SeatID
}

func NewDealStateMachine(matchCtx *domain.MatchCtx, eventBus *event.EventBus) *DealStateMachine {
	return &DealStateMachine{
		currentPhase: PhaseIdle,
		matchCtx:     matchCtx,
		eventBus:     eventBus,
	}
}

func (sm *DealStateMachine) GetCurrentPhase() DealPhase {
	return sm.currentPhase
}

func (sm *DealStateMachine) GetMatchCtx() *domain.MatchCtx {
	return sm.matchCtx
}

func (sm *DealStateMachine) GetDealCtx() *domain.DealCtx {
	return sm.dealCtx
}

func (sm *DealStateMachine) GetTrickCtx() *domain.TrickCtx {
	return sm.trickCtx
}

func (sm *DealStateMachine) StartDeal(dealNumber int, firstPlayer domain.SeatID) error {
	if sm.currentPhase != PhaseIdle {
		return fmt.Errorf("cannot start deal from phase %s", sm.currentPhase.String())
	}
	
	// Create deal context without trump - trump will be determined in P2 phase
	sm.dealCtx = domain.NewDealCtx(dealNumber, domain.Two, firstPlayer) // Temporary trump
	sm.currentPhase = PhaseCreated
	
	sm.eventBus.Publish(event.NewDealStartedEvent(
		sm.matchCtx.ID,
		dealNumber,
		domain.Two, // Temporary trump
		firstPlayer,
	))
	
	return nil
}

func (sm *DealStateMachine) DealCards() error {
	if sm.currentPhase != PhaseCreated {
		return fmt.Errorf("cannot deal cards from phase %s", sm.currentPhase.String())
	}
	
	sm.deck = domain.NewDeckWithSeed(sm.matchCtx.Seed)
	sm.deck.Shuffle()
	
	// P1 Step 2: Select starting card for first deal
	if sm.dealCtx.IsFirstDeal {
		// Randomly select a starting card from the deck
		allCards := sm.deck.Cards
		startingCardIndex := sm.matchCtx.Seed % int64(len(allCards))
		sm.startingCard = &allCards[startingCardIndex]
	}
	
	// P1 Step 3: Deal cards to each player (27 cards each)
	hands := sm.deck.DealToHands(4)
	handMap := make(map[domain.SeatID][]domain.Card)
	
	for i, hand := range hands {
		seat := domain.SeatID(i)
		player := sm.matchCtx.GetPlayer(seat)
		if player != nil {
			player.ClearHand()
			player.AddCards(hand)
			handMap[seat] = hand
			
			// P1 Step 4: Record starting card holder for first deal
			if sm.dealCtx.IsFirstDeal && sm.startingCard != nil {
				for _, card := range hand {
					if card.Rank == sm.startingCard.Rank && card.Suit == sm.startingCard.Suit {
						sm.startingCardHolder = seat
						break
					}
				}
			}
		}
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateDealt)
	sm.currentPhase = PhaseCardsDealt
	
	sm.eventBus.Publish(event.NewCardsDealtEvent(
		sm.matchCtx.ID,
		handMap,
	))
	
	return nil
}

// DetermineTrump implements P2 phase - Determine Level & Trump
func (sm *DealStateMachine) DetermineTrump() error {
	if sm.currentPhase != PhaseCardsDealt {
		return fmt.Errorf("cannot determine trump from phase %s", sm.currentPhase.String())
	}

	// P2 Step 1: Read previous deal winner team's level, default to 2 if none
	var currentLevel domain.Rank = domain.Two
	if sm.dealCtx.DealNumber > 1 {
		// Get previous deal winner team's current level
		// For now, we'll use team1's level as placeholder
		// This should be updated based on previous deal results
		team := sm.matchCtx.GetTeam(domain.TeamEastWest)
		if team != nil {
			currentLevel = team.Level
		}
	}

	// P2 Step 2: Set 8 cards equal to Level as Trump
	trump := currentLevel

	// P2 Step 3: Write Level & Trump to Deal-ctx
	sm.dealCtx = sm.dealCtx.WithTrump(trump)
	sm.dealCtx = sm.dealCtx.WithCurrentLevel(currentLevel)

	sm.currentPhase = PhaseTrumpDecision

	sm.eventBus.Publish(event.NewTrumpDeterminedEvent(
		sm.matchCtx.ID,
		currentLevel,
		trump,
	))

	return nil
}

func (sm *DealStateMachine) StartTribute() error {
	if sm.currentPhase != PhaseTrumpDecision {
		return fmt.Errorf("cannot start tribute from phase %s", sm.currentPhase.String())
	}
	
	if sm.dealCtx.IsFirstDeal {
		return sm.skipTribute()
	}
	
	// 计算每个玩家的大王数量
	playerBigJokers := make(map[domain.SeatID]int)
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		player := sm.matchCtx.GetPlayer(seat)
		if player != nil {
			playerBigJokers[seat] = domain.CountBigJokers(player.GetHand())
		}
	}
	
	// 初始化贡牌系统
	sm.dealCtx = sm.dealCtx.InitializeTribute(playerBigJokers)
	
	if sm.dealCtx.TributeInfo.HasImmunity {
		// 有免疫，直接跳过贡牌
		return sm.skipTribute()
	}
	
	// 转换贡牌要求格式用于事件
	tributeRequirements := make(map[domain.SeatID]int)
	for from := range sm.dealCtx.TributeInfo.TributeRequests {
		tributeRequirements[from] = 1 // 每人贡1张牌
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateTribute)
	sm.currentPhase = PhaseTribute
	
	sm.eventBus.Publish(event.NewTributeRequestedEvent(
		sm.matchCtx.ID,
		tributeRequirements,
	))
	
	return nil
}

func (sm *DealStateMachine) skipTribute() error {
	sm.dealCtx = sm.dealCtx.WithTributeGiven(true)
	return sm.StartFirstPlay()
}

func (sm *DealStateMachine) GiveTribute(from, to domain.SeatID, cards []domain.Card) error {
	if sm.currentPhase != PhaseTribute {
		return fmt.Errorf("cannot give tribute from phase %s", sm.currentPhase.String())
	}
	
	// 验证贡牌数量
	if len(cards) != 1 {
		return fmt.Errorf("must tribute exactly one card")
	}
	
	card := cards[0]
	
	fromPlayer := sm.matchCtx.GetPlayer(from)
	toPlayer := sm.matchCtx.GetPlayer(to)
	
	if fromPlayer == nil || toPlayer == nil {
		return fmt.Errorf("invalid player seats")
	}
	
	// 验证贡牌关系
	expectedTo, exists := sm.dealCtx.TributeInfo.TributeRequests[from]
	if !exists {
		return fmt.Errorf("player %s is not required to give tribute", from.String())
	}
	if expectedTo != to {
		return fmt.Errorf("player %s should give tribute to %s, not %s", from.String(), expectedTo.String(), to.String())
	}
	
	// 验证贡牌是否符合规则（除了红桃trump外最大的牌）
	if err := domain.ValidateTributeCard(fromPlayer.GetHand(), card, sm.dealCtx.Trump); err != nil {
		return fmt.Errorf("invalid tribute card: %w", err)
	}
	
	// 执行贡牌
	fromPlayer.RemoveCards(cards)
	toPlayer.AddCards(cards)
	
	// 记录贡牌
	sm.dealCtx.TributeCards[from] = cards
	sm.dealCtx.TributeInfo.GivenTributes[from] = card
	
	sm.eventBus.Publish(event.NewTributeGivenEvent(
		sm.matchCtx.ID,
		from,
		to,
		cards,
	))
	
	// 检查是否所有贡牌都已完成
	if sm.dealCtx.TributeInfo.IsTributeComplete() {
		sm.dealCtx.TributeInfo.Phase = domain.TributePhaseGiving
		
		// Double Down场景需要进入选择阶段
		if sm.dealCtx.TributeInfo.Scenario == domain.TributeScenarioDoubleDown {
			return sm.StartTributeSelection()
		}
		
		return sm.StartReturnTribute()
	}
	
	return nil
}

// StartTributeSelection 开始Double Down贡牌选择阶段
func (sm *DealStateMachine) StartTributeSelection() error {
	if sm.currentPhase != PhaseTribute {
		return fmt.Errorf("cannot start tribute selection from phase %s", sm.currentPhase.String())
	}
	
	if sm.dealCtx.TributeInfo.Scenario != domain.TributeScenarioDoubleDown {
		return fmt.Errorf("tribute selection only available for Double Down scenario")
	}
	
	// 准备选择阶段
	sm.dealCtx.TributeInfo.PrepareDoubleDownSelection(sm.dealCtx.LastRankings)
	sm.currentPhase = PhaseTributeSelection
	
	// 发布选择请求事件
	first := sm.dealCtx.LastRankings[0]
	sm.eventBus.Publish(event.NewTributeSelectionRequestedEvent(
		sm.matchCtx.ID,
		first,
		sm.dealCtx.TributeInfo.AvailableCards,
	))
	
	return nil
}

// SelectTributeCard Player 1在Double Down场景中选择贡牌
func (sm *DealStateMachine) SelectTributeCard(giver domain.SeatID) error {
	if sm.currentPhase != PhaseTributeSelection {
		return fmt.Errorf("cannot select tribute card from phase %s", sm.currentPhase.String())
	}
	
	if sm.dealCtx.TributeInfo.Scenario != domain.TributeScenarioDoubleDown {
		return fmt.Errorf("tribute selection only available for Double Down scenario")
	}
	
	// 获取选择前的信息，用于事件发布
	selectedCard := sm.dealCtx.TributeInfo.AvailableCards[giver]
	first := sm.dealCtx.LastRankings[0]
	second := sm.dealCtx.LastRankings[1]
	third := sm.dealCtx.LastRankings[2]
	fourth := sm.dealCtx.LastRankings[3]
	
	var remainingGiver domain.SeatID
	if giver == third {
		remainingGiver = fourth
	} else {
		remainingGiver = third
	}
	remainingCard := sm.dealCtx.TributeInfo.AvailableCards[remainingGiver]
	
	// 执行选择
	err := sm.dealCtx.TributeInfo.SelectTributeCardForDoubleDown(giver, sm.dealCtx.LastRankings)
	if err != nil {
		return fmt.Errorf("failed to select tribute card: %w", err)
	}
	
	// 实际进行卡牌交换
	firstPlayer := sm.matchCtx.GetPlayer(first)
	secondPlayer := sm.matchCtx.GetPlayer(second)
	
	if firstPlayer != nil {
		firstPlayer.AddCards([]domain.Card{selectedCard})
	}
	if secondPlayer != nil {
		secondPlayer.AddCards([]domain.Card{remainingCard})
	}
	
	// 发布选择完成事件
	sm.eventBus.Publish(event.NewTributeCardSelectedEvent(
		sm.matchCtx.ID,
		first,
		giver,
		selectedCard,
		second,
		remainingCard,
	))
	
	// 转到还贡阶段
	return sm.StartReturnTribute()
}

// StartReturnTribute 开始还贡阶段
func (sm *DealStateMachine) StartReturnTribute() error {
	if sm.currentPhase != PhaseTribute && sm.currentPhase != PhaseTributeSelection {
		return fmt.Errorf("cannot start return tribute from phase %s", sm.currentPhase.String())
	}
	
	if sm.dealCtx.TributeInfo.HasImmunity || len(sm.dealCtx.TributeInfo.ReturnRequests) == 0 {
		// 无需还贡，直接开始首次出牌
		sm.dealCtx.TributeInfo.Phase = domain.TributePhaseCompleted
		sm.dealCtx = sm.dealCtx.WithTributeGiven(true)
		return sm.StartFirstPlay()
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateReturnTribute)
	sm.currentPhase = PhaseReturnTribute
	sm.dealCtx.TributeInfo.Phase = domain.TributePhaseReturning
	
	return nil
}

// GiveReturnTribute 执行还贡
func (sm *DealStateMachine) GiveReturnTribute(from, to domain.SeatID, cards []domain.Card) error {
	if sm.currentPhase != PhaseReturnTribute {
		return fmt.Errorf("cannot give return tribute from phase %s", sm.currentPhase.String())
	}
	
	// 验证还贡数量
	if len(cards) != 1 {
		return fmt.Errorf("must return exactly one card")
	}
	
	card := cards[0]
	
	fromPlayer := sm.matchCtx.GetPlayer(from)
	toPlayer := sm.matchCtx.GetPlayer(to)
	
	if fromPlayer == nil || toPlayer == nil {
		return fmt.Errorf("invalid player seats")
	}
	
	// 验证还贡关系
	expectedTo, exists := sm.dealCtx.TributeInfo.ReturnRequests[from]
	if !exists {
		return fmt.Errorf("player %s is not required to give return tribute", from.String())
	}
	if expectedTo != to {
		return fmt.Errorf("player %s should give return tribute to %s, not %s", from.String(), expectedTo.String(), to.String())
	}
	
	// 验证还贡牌是否符合规则（点数<=10）
	if !domain.IsValidReturnTributeCard(fromPlayer.GetHand(), card) {
		return fmt.Errorf("invalid return tribute card: must be <= 10 points")
	}
	
	// 执行还贡
	fromPlayer.RemoveCards(cards)
	toPlayer.AddCards(cards)
	
	// 记录还贡
	sm.dealCtx.TributeInfo.ReturnedTributes[from] = card
	
	sm.eventBus.Publish(event.NewTributeGivenEvent(
		sm.matchCtx.ID,
		from,
		to,
		cards,
	))
	
	// 检查是否所有还贡都已完成
	if sm.dealCtx.TributeInfo.IsReturnComplete() {
		sm.dealCtx.TributeInfo.Phase = domain.TributePhaseCompleted
		sm.dealCtx = sm.dealCtx.WithTributeGiven(true)
		return sm.StartFirstPlay()
	}
	
	return nil
}

func (sm *DealStateMachine) StartFirstPlay() error {
	if sm.currentPhase != PhaseTribute && sm.currentPhase != PhaseReturnTribute && sm.currentPhase != PhaseTrumpDecision {
		return fmt.Errorf("cannot start first play from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateFirstPlay)
	sm.currentPhase = PhaseFirstPlay
	
	sm.trickCtx = domain.NewTrickCtx(1, sm.dealCtx.FirstPlayer)
	
	return nil
}

func (sm *DealStateMachine) TransitionToInProgress() error {
	if sm.currentPhase != PhaseFirstPlay {
		return fmt.Errorf("cannot transition to in progress from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateInProgress)
	sm.currentPhase = PhaseInProgress
	
	return nil
}

func (sm *DealStateMachine) StartNewTrick(startPlayer domain.SeatID) error {
	if sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot start new trick from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithTrickCount(sm.dealCtx.TrickCount + 1)
	sm.trickCtx = domain.NewTrickCtx(sm.dealCtx.TrickCount, startPlayer)
	
	return nil
}

func (sm *DealStateMachine) PlayCards(seat domain.SeatID, cards []domain.Card) error {
	if sm.currentPhase != PhaseFirstPlay && sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot play cards from phase %s", sm.currentPhase.String())
	}
	
	if sm.trickCtx.CurrentPlayer != seat {
		return fmt.Errorf("not player's turn")
	}
	
	player := sm.matchCtx.GetPlayer(seat)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	
	if !player.HasCards(cards) {
		return fmt.Errorf("player does not have required cards")
	}
	
	cardGroup := domain.NewCardGroup(cards)
	if !cardGroup.IsValid() {
		return fmt.Errorf("invalid card combination")
	}
	
	if !domain.CanFollow(cardGroup, sm.trickCtx.LastPlay, sm.dealCtx.Trump) {
		return fmt.Errorf("cannot beat current play")
	}
	
	player.RemoveCards(cards)
	
	sm.trickCtx = sm.trickCtx.WithLastPlay(cardGroup, seat)
	sm.trickCtx = sm.trickCtx.WithCurrentPlayer(seat.Next())
	
	trickPlay := domain.TrickPlay{
		Player:    seat,
		Cards:     cards,
		CardGroup: cardGroup,
		Timestamp: time.Now(),
	}
	sm.trickCtx = sm.trickCtx.WithPlayHistory(trickPlay)
	
	sm.eventBus.Publish(event.NewCardsPlayedEvent(
		sm.matchCtx.ID,
		seat,
		cards,
		cardGroup,
	))
	
	if sm.currentPhase == PhaseFirstPlay {
		sm.TransitionToInProgress()
	}
	
	if player.IsHandEmpty() {
		return sm.handlePlayerFinished(seat)
	}
	
	return nil
}

func (sm *DealStateMachine) Pass(seat domain.SeatID) error {
	if sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot pass from phase %s", sm.currentPhase.String())
	}
	
	if sm.trickCtx.CurrentPlayer != seat {
		return fmt.Errorf("not player's turn")
	}
	
	sm.trickCtx = sm.trickCtx.WithPlayerPassed(seat)
	sm.trickCtx = sm.trickCtx.WithCurrentPlayer(seat.Next())
	
	sm.eventBus.Publish(event.NewPlayerPassedEvent(
		sm.matchCtx.ID,
		seat,
	))
	
	if sm.trickCtx.ShouldFinish() {
		return sm.finishTrick()
	}
	
	return nil
}

func (sm *DealStateMachine) finishTrick() error {
	winner := sm.trickCtx.LastPlayer
	sm.trickCtx = sm.trickCtx.WithWinner(winner)
	
	sm.eventBus.Publish(event.NewTrickWonEvent(
		sm.matchCtx.ID,
		winner,
		sm.trickCtx.TrickNumber,
	))
	
	if sm.shouldFinishDeal() {
		return sm.finishDeal()
	}
	
	return sm.StartNewTrick(winner)
}

func (sm *DealStateMachine) handlePlayerFinished(seat domain.SeatID) error {
	position := len(sm.dealCtx.RankList) + 1
	sm.dealCtx = sm.dealCtx.WithRankList(append(sm.dealCtx.RankList, seat))
	
	sm.eventBus.Publish(event.NewPlayerFinishedEvent(
		sm.matchCtx.ID,
		seat,
		position,
	))
	
	if sm.shouldFinishDeal() {
		return sm.finishDeal()
	}
	
	return nil
}

func (sm *DealStateMachine) shouldFinishDeal() bool {
	return len(sm.dealCtx.RankList) >= 3
}

func (sm *DealStateMachine) finishDeal() error {
	sm.currentPhase = PhaseRankList
	
	winnerTeam := sm.determineWinnerTeam()
	
	sm.eventBus.Publish(event.NewDealEndedEvent(
		sm.matchCtx.ID,
		sm.dealCtx.DealNumber,
		sm.dealCtx.RankList,
		winnerTeam,
	))
	
	if sm.shouldFinishMatch() {
		return sm.finishMatch(winnerTeam)
	}
	
	sm.currentPhase = PhaseFinished
	return nil
}

func (sm *DealStateMachine) finishMatch(winnerTeam domain.TeamID) error {
	sm.currentPhase = PhaseFinished
	
	finalScore := make(map[domain.TeamID]int)
	finalScore[winnerTeam] = 1
	finalScore[winnerTeam.OpposingTeam()] = 0
	
	sm.eventBus.Publish(event.NewMatchEndedEvent(
		sm.matchCtx.ID,
		winnerTeam,
		finalScore,
	))
	
	sm.matchCtx = sm.matchCtx.WithWinner(winnerTeam)
	
	return nil
}

func (sm *DealStateMachine) determineWinnerTeam() domain.TeamID {
	if len(sm.dealCtx.RankList) >= 2 {
		first := sm.dealCtx.RankList[0]
		second := sm.dealCtx.RankList[1]
		
		if domain.GetTeamFromSeat(first) == domain.GetTeamFromSeat(second) {
			return domain.GetTeamFromSeat(first)
		}
	}
	
	return domain.GetTeamFromSeat(sm.dealCtx.RankList[0])
}

func (sm *DealStateMachine) shouldFinishMatch() bool {
	return sm.dealCtx.CurrentLevel >= domain.Ace
}

// GetTributeCardOptions 获取贡牌选项（调试用）
func (sm *DealStateMachine) GetTributeCardOptions(seat domain.SeatID) []domain.Card {
	player := sm.matchCtx.GetPlayer(seat)
	if player == nil {
		return nil
	}
	
	return domain.GetTributeCardCandidates(player.GetHand(), sm.dealCtx.Trump)
}

// GetReturnTributeCardOptions 获取还贡选项
func (sm *DealStateMachine) GetReturnTributeCardOptions(seat domain.SeatID) []domain.Card {
	player := sm.matchCtx.GetPlayer(seat)
	if player == nil {
		return nil
	}
	
	return domain.GetReturnTributeCardCandidates(player.GetHand())
}

func (sm *DealStateMachine) Reset() {
	sm.currentPhase = PhaseIdle
	sm.dealCtx = nil
	sm.trickCtx = nil
	sm.deck = nil
	sm.startingCard = nil
	sm.startingCardHolder = domain.SeatEast
}