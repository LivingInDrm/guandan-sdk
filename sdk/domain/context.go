package domain

import (
	"time"
)

type MatchID string

type MatchState int

const (
	MatchStateCreated MatchState = iota
	MatchStateInProgress
	MatchStateFinished
)

func (s MatchState) String() string {
	switch s {
	case MatchStateCreated:
		return "Created"
	case MatchStateInProgress:
		return "InProgress"
	case MatchStateFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type MatchCtx struct {
	ID          MatchID
	State       MatchState
	Players     PlayerArray
	Teams       [2]*Team
	StartTime   time.Time
	EndTime     *time.Time
	CurrentDeal int
	MaxDeals    int
	Winner      *TeamID
	Seed        int64
}

func NewMatchCtx(id MatchID, players []*Player, seed int64) *MatchCtx {
	playerArray := NewPlayerArray()
	teams := [2]*Team{
		NewTeam(TeamEastWest),
		NewTeam(TeamSouthNorth),
	}
	
	for _, player := range players {
		playerArray.Set(player.SeatID, player)
		teams[player.TeamID].AddPlayer(player)
	}
	
	return &MatchCtx{
		ID:          id,
		State:       MatchStateCreated,
		Players:     playerArray,
		Teams:       teams,
		StartTime:   time.Now(),
		CurrentDeal: 0,
		MaxDeals:    0,
		Seed:        seed,
	}
}

func (m *MatchCtx) GetPlayer(seat SeatID) *Player {
	return m.Players.Get(seat)
}

func (m *MatchCtx) GetTeam(teamID TeamID) *Team {
	return m.Teams[teamID]
}

func (m *MatchCtx) IsFinished() bool {
	return m.State == MatchStateFinished
}

func (m *MatchCtx) WithState(state MatchState) *MatchCtx {
	newCtx := *m
	newCtx.State = state
	return &newCtx
}

func (m *MatchCtx) WithCurrentDeal(deal int) *MatchCtx {
	newCtx := *m
	newCtx.CurrentDeal = deal
	return &newCtx
}

func (m *MatchCtx) WithWinner(winner TeamID) *MatchCtx {
	newCtx := *m
	newCtx.Winner = &winner
	now := time.Now()
	newCtx.EndTime = &now
	newCtx.State = MatchStateFinished
	return &newCtx
}

type DealState int

const (
	DealStateCreated DealState = iota
	DealStateDealt
	DealStateTribute
	DealStateReturnTribute
	DealStateFirstPlay
	DealStateInProgress
	DealStateFinished
)

func (s DealState) String() string {
	switch s {
	case DealStateCreated:
		return "Created"
	case DealStateDealt:
		return "Dealt"
	case DealStateTribute:
		return "Tribute"
	case DealStateReturnTribute:
		return "ReturnTribute"
	case DealStateFirstPlay:
		return "FirstPlay"
	case DealStateInProgress:
		return "InProgress"
	case DealStateFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type DealCtx struct {
	DealNumber     int
	State          DealState
	Trump          Rank
	CurrentLevel   Rank
	StartTime      time.Time
	EndTime        *time.Time
	FirstPlayer    SeatID
	RankList       []SeatID
	TrickCount     int
	IsFirstDeal    bool
	TributeGiven   bool
	TributeCards   map[SeatID][]Card
	LastRankings   []SeatID      // 上局排名，用于计算贡牌
	TributeInfo    *TributeInfo  // 新的贡牌信息
}

func NewDealCtx(dealNumber int, trump Rank, firstPlayer SeatID) *DealCtx {
	return &DealCtx{
		DealNumber:   dealNumber,
		State:        DealStateCreated,
		Trump:        trump,
		CurrentLevel: trump,
		StartTime:    time.Now(),
		FirstPlayer:  firstPlayer,
		RankList:     make([]SeatID, 0, 4),
		TrickCount:   0,
		IsFirstDeal:  dealNumber == 1,
		TributeGiven: false,
		TributeCards: make(map[SeatID][]Card),
		LastRankings: nil,
		TributeInfo:  nil,
	}
}

// NewDealCtxWithHistory 创建带历史排名的DealCtx
func NewDealCtxWithHistory(dealNumber int, trump Rank, firstPlayer SeatID, lastRankings []SeatID) *DealCtx {
	ctx := NewDealCtx(dealNumber, trump, firstPlayer)
	ctx.LastRankings = lastRankings
	return ctx
}

// NewDealCtxWithAutoFirstPlayer 创建自动确定首出者的DealCtx
func NewDealCtxWithAutoFirstPlayer(matchCtx *MatchCtx, dealNumber int, trump Rank, lastRankings []SeatID, startingCardHolder SeatID) *DealCtx {
	// 创建临时的DealCtx来进行首出者判定
	tempCtx := &DealCtx{
		DealNumber:   dealNumber,
		IsFirstDeal:  dealNumber == 1,
		LastRankings: lastRankings,
		Trump:        trump,
	}
	
	// 确定首出者
	firstPlayer := DetermineFirstPlayer(matchCtx, tempCtx, startingCardHolder)
	
	// 创建最终的DealCtx
	ctx := NewDealCtx(dealNumber, trump, firstPlayer)
	ctx.LastRankings = lastRankings
	return ctx
}

func (d *DealCtx) WithState(state DealState) *DealCtx {
	newCtx := *d
	newCtx.State = state
	return &newCtx
}

func (d *DealCtx) WithTrickCount(count int) *DealCtx {
	newCtx := *d
	newCtx.TrickCount = count
	return &newCtx
}

func (d *DealCtx) WithRankList(rankList []SeatID) *DealCtx {
	newCtx := *d
	newCtx.RankList = make([]SeatID, len(rankList))
	copy(newCtx.RankList, rankList)
	return &newCtx
}

func (d *DealCtx) WithTributeGiven(given bool) *DealCtx {
	newCtx := *d
	newCtx.TributeGiven = given
	return &newCtx
}

func (d *DealCtx) WithTrump(trump Rank) *DealCtx {
	newCtx := *d
	newCtx.Trump = trump
	return &newCtx
}

func (d *DealCtx) WithCurrentLevel(level Rank) *DealCtx {
	newCtx := *d
	newCtx.CurrentLevel = level
	return &newCtx
}

func (d *DealCtx) WithEndTime(endTime time.Time) *DealCtx {
	newCtx := *d
	newCtx.EndTime = &endTime
	newCtx.State = DealStateFinished
	return &newCtx
}

func (d *DealCtx) AddToRankList(seat SeatID) *DealCtx {
	newCtx := *d
	newCtx.RankList = make([]SeatID, len(d.RankList)+1)
	copy(newCtx.RankList, d.RankList)
	newCtx.RankList[len(d.RankList)] = seat
	return &newCtx
}

func (d *DealCtx) IsFinished() bool {
	return d.State == DealStateFinished
}

func (d *DealCtx) GetRankPosition(seat SeatID) int {
	for i, s := range d.RankList {
		if s == seat {
			return i + 1
		}
	}
	return 0
}

// WithTributeInfo 设置贡牌信息
func (d *DealCtx) WithTributeInfo(tributeInfo *TributeInfo) *DealCtx {
	newCtx := *d
	newCtx.TributeInfo = tributeInfo
	return &newCtx
}

// WithLastRankings 设置上局排名
func (d *DealCtx) WithLastRankings(rankings []SeatID) *DealCtx {
	newCtx := *d
	newCtx.LastRankings = make([]SeatID, len(rankings))
	copy(newCtx.LastRankings, rankings)
	return &newCtx
}

// InitializeTribute 初始化贡牌系统
func (d *DealCtx) InitializeTribute(playerBigJokers map[SeatID]int) *DealCtx {
	if d.IsFirstDeal {
		// 首局无需贡牌
		return d.WithTributeInfo(NewTributeInfo(TributeScenarioNone, false))
	}

	scenario := DetermineTributeScenario(d.LastRankings)
	hasImmunity := CheckTributeImmunity(scenario, playerBigJokers, d.LastRankings)
	
	tributeInfo := NewTributeInfo(scenario, hasImmunity)
	
	if !hasImmunity {
		// 设置贡牌要求
		tributeRequests := CalculateTributeRequirements(scenario, d.LastRankings)
		for from, to := range tributeRequests {
			tributeInfo.TributeRequests[from] = to
		}
		
		// 设置还贡要求
		returnRequests := CalculateReturnRequirements(scenario, d.LastRankings)
		for from, to := range returnRequests {
			tributeInfo.ReturnRequests[from] = to
		}
		
		tributeInfo.Phase = TributePhaseRequested
	} else {
		tributeInfo.Phase = TributePhaseCompleted
	}
	
	return d.WithTributeInfo(tributeInfo)
}

type TrickState int

const (
	TrickStateActive TrickState = iota
	TrickStateFinished
)

func (s TrickState) String() string {
	switch s {
	case TrickStateActive:
		return "Active"
	case TrickStateFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type TrickCtx struct {
	TrickNumber   int
	State         TrickState
	StartPlayer   SeatID
	CurrentPlayer SeatID
	LastPlay      *CardGroup
	LastPlayer    SeatID
	PassedPlayers map[SeatID]bool
	PlayHistory   []TrickPlay
	Winner        SeatID
}

type TrickPlay struct {
	Player    SeatID
	Cards     []Card
	CardGroup *CardGroup
	Timestamp time.Time
}

func NewTrickCtx(trickNumber int, startPlayer SeatID) *TrickCtx {
	return &TrickCtx{
		TrickNumber:   trickNumber,
		State:         TrickStateActive,
		StartPlayer:   startPlayer,
		CurrentPlayer: startPlayer,
		PassedPlayers: make(map[SeatID]bool),
		PlayHistory:   make([]TrickPlay, 0),
	}
}

func (t *TrickCtx) WithCurrentPlayer(player SeatID) *TrickCtx {
	newCtx := *t
	newCtx.CurrentPlayer = player
	return &newCtx
}

func (t *TrickCtx) WithLastPlay(cardGroup *CardGroup, player SeatID) *TrickCtx {
	newCtx := *t
	newCtx.LastPlay = cardGroup
	newCtx.LastPlayer = player
	return &newCtx
}

func (t *TrickCtx) WithPlayerPassed(player SeatID) *TrickCtx {
	newCtx := *t
	newCtx.PassedPlayers = make(map[SeatID]bool)
	for k, v := range t.PassedPlayers {
		newCtx.PassedPlayers[k] = v
	}
	newCtx.PassedPlayers[player] = true
	return &newCtx
}

func (t *TrickCtx) WithPlayHistory(play TrickPlay) *TrickCtx {
	newCtx := *t
	newCtx.PlayHistory = make([]TrickPlay, len(t.PlayHistory)+1)
	copy(newCtx.PlayHistory, t.PlayHistory)
	newCtx.PlayHistory[len(t.PlayHistory)] = play
	return &newCtx
}

func (t *TrickCtx) WithWinner(winner SeatID) *TrickCtx {
	newCtx := *t
	newCtx.Winner = winner
	newCtx.State = TrickStateFinished
	return &newCtx
}

func (t *TrickCtx) HasPlayerPassed(player SeatID) bool {
	return t.PassedPlayers[player]
}

func (t *TrickCtx) GetActivePlayerCount() int {
	activeCount := 4
	for _, passed := range t.PassedPlayers {
		if passed {
			activeCount--
		}
	}
	return activeCount
}

func (t *TrickCtx) IsFinished() bool {
	return t.State == TrickStateFinished
}

func (t *TrickCtx) GetNextPlayer() SeatID {
	return t.CurrentPlayer.Next()
}

func (t *TrickCtx) ShouldFinish() bool {
	return t.GetActivePlayerCount() <= 1
}

// DetermineFirstPlayer 根据游戏规则确定本Deal的首出者
func DetermineFirstPlayer(matchCtx *MatchCtx, dealCtx *DealCtx, startingCardHolder SeatID) SeatID {
	// 首Deal：持有Starting Card者先出
	if dealCtx.IsFirstDeal {
		return startingCardHolder
	}
	
	// 非首Deal：根据上一Deal结果和是否有Tribute Immunity判定
	if dealCtx.LastRankings == nil || len(dealCtx.LastRankings) != 4 {
		// 数据不完整，默认使用East
		return SeatEast
	}
	
	// 获取上一Deal的排名
	first := dealCtx.LastRankings[0]
	fourth := dealCtx.LastRankings[3]
	third := dealCtx.LastRankings[2]
	
	// 判断贡牌场景
	scenario := DetermineTributeScenario(dealCtx.LastRankings)
	
	// 检查是否有Tribute Immunity
	hasImmunity := false
	if dealCtx.TributeInfo != nil {
		hasImmunity = dealCtx.TributeInfo.HasImmunity
	}
	
	switch scenario {
	case TributeScenarioDoubleDown:
		if hasImmunity {
			// 有Immunity：上轮Deal第1名
			return first
		} else {
			// 无Immunity：首选贡牌较大者；若两张贡牌一样大，则选择上轮Deal的第一名的顺时针下家
			return determineDoubleDownFirstPlayer(dealCtx, first)
		}
		
	case TributeScenarioSingleLast:
		if hasImmunity {
			// 有Immunity：上轮Deal第1名
			return first
		} else {
			// 无Immunity：上轮Deal的第4名
			return fourth
		}
		
	case TributeScenarioPartnerLast:
		if hasImmunity {
			// 有Immunity：上轮Deal第1名
			return first
		} else {
			// 无Immunity：上轮Deal的第3名
			return third
		}
		
	default:
		// 其他情况默认使用第一名
		return first
	}
}

// determineDoubleDownFirstPlayer 确定DoubleDown场景下的首出者
func determineDoubleDownFirstPlayer(dealCtx *DealCtx, first SeatID) SeatID {
	if dealCtx.TributeInfo == nil || len(dealCtx.TributeInfo.GivenTributes) == 0 {
		// 没有贡牌信息，使用第一名的顺时针下家
		return first.Next()
	}
	
	// 获取两张贡牌
	third := dealCtx.LastRankings[2]
	fourth := dealCtx.LastRankings[3]
	
	thirdCard, thirdExists := dealCtx.TributeInfo.GivenTributes[third]
	fourthCard, fourthExists := dealCtx.TributeInfo.GivenTributes[fourth]
	
	// 如果贡牌信息不完整，使用第一名的顺时针下家
	if !thirdExists || !fourthExists {
		return first.Next()
	}
	
	// 比较两张贡牌的大小
	trump := dealCtx.Trump
	thirdValue := getTributeCardValue(thirdCard, trump)
	fourthValue := getTributeCardValue(fourthCard, trump)
	
	if thirdValue > fourthValue {
		// 第三名贡牌更大，选择第三名的贡牌接收者
		return getDoubleDownReceiver(dealCtx, third)
	} else if fourthValue > thirdValue {
		// 第四名贡牌更大，选择第四名的贡牌接收者
		return getDoubleDownReceiver(dealCtx, fourth)
	} else {
		// 两张贡牌一样大，选择第一名的顺时针下家
		return first.Next()
	}
}

// getDoubleDownReceiver 获取DoubleDown场景下的实际接收者
func getDoubleDownReceiver(dealCtx *DealCtx, giver SeatID) SeatID {
	// 如果有ActualReceivers信息，使用实际接收者
	if dealCtx.TributeInfo != nil && dealCtx.TributeInfo.ActualReceivers != nil {
		if receiver, exists := dealCtx.TributeInfo.ActualReceivers[giver]; exists {
			return receiver
		}
	}
	
	// 否则使用默认映射：第三名->第一名，第四名->第二名
	if len(dealCtx.LastRankings) >= 4 {
		first := dealCtx.LastRankings[0]
		second := dealCtx.LastRankings[1]
		third := dealCtx.LastRankings[2]
		
		if giver == third {
			return first
		} else {
			return second
		}
	}
	
	return SeatEast // 默认值
}