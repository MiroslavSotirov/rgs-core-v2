package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strconv"
	"strings"
)

const DefaultSchema = "http://json-schemas.maverick.com/general/maverick-v1.json"

// OKResponse ...
type OKResponse struct {
	HTTPStatusCode int `json:"-"` // http response status code
}

// Render renders OKResponse struct
func (e *OKResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// OKNoContent returns no content http status
var OKNoContent = &OKResponse{HTTPStatusCode: 204}

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int    `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render error response
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest Render invalid request
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ErrRender render all sorts of error
func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// ErrBadRequestRender Bad Request error render
func ErrBadRequestRender(err *rgserror.RGSError) render.Renderer {
	statusCode := http.StatusBadRequest
	statusText := "Bad Request"
	if err.ErrCode == 450 { // casting int to int64 might cause a bug?
		statusCode = 450
		statusText = "Insufficient Fund"
	}

	return &ErrResponse{
		HTTPStatusCode: statusCode,
		StatusText:     statusText,
		AppCode:        err.ErrCode,
		ErrorText:      err.Error(),
	}
}

// ErrNotFound returns 404
var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

//// ErrUnauthorized returns 401
//var ErrUnauthorized = &ErrResponse{HTTPStatusCode: 401, StatusText: "UnAuthorized. Login required."}
//
//// ErrPrecondition returns 412
//var ErrPrecondition = &ErrResponse{HTTPStatusCode: 412, StatusText: "Precondition required."}

var ErrInternalServerError = &ErrResponse{HTTPStatusCode: 500, StatusText: "Internal server error", AppCode: rgserror.ErrInternalServerError.ErrCode, ErrorText: rgserror.ErrInternalServerError.Error()}

// SystemInit Response
type SystemInit struct {
	Message string `json:"message"`
}

// Render system init message
func (si SystemInit) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GameInitResponse reponse
type GameInitResponse struct {
	Schema     string                 `json:"$schema,omitempty"`
	Balance    float32                `json:"balance"`
	Currency   string                 `json:"currency"`
	Parameters ParameterResponse      `json:"parameters"`
	Player     PlayerResponse         `json:"player"`
	Definition GameDefinitionResponse `json:"definition"`
	Links      []LinkResponse         `json:"_links"`
	Forms      FormResponse           `json:"_forms"`
	Gameplay   GameplayResponse       `json:"gameplay"`
}

// GameDefinitionResponse ...
type GameDefinitionResponse struct {
	WinLineDefs [][]int                      `json:"winlineDefs"`
	WinDefs     map[string]WinResponse       `json:"winDefs"`
	ReelSetDefs map[string]EngineDefResponse `json:"reelsetDefs"`
	ReelsInit   [][]int                      `json:"reelsInit"`
	PrizeDefs   map[string]PrizeDefResponse  `json:"prizeDefs,omitempty"`
	SymbolDefs  map[int]SymbolDefResponse    `json:"symbolDefs"`
}

type SymbolDefResponse struct {
	IsWild      bool   `json:"isWild"`
	DisplayName string `json:"displayName"`
}

type EngineDefResponse interface {
	//MarshalJSON() ([]byte, error)
}
type EngineIDefResponse struct {
	Reels [][]int `json:"reels"`
	Type  string  `json:"type"`
}
type EngineIIDefResponse struct {
	Reels      [][]int               `json:"reels"`
	Type       string                `json:"type"`
	NoWilds    *FeatureReelsResponse `json:"0,omitempty"`
	OneWild    *FeatureReelsResponse `json:"1,omitempty"`
	TwoWilds   *FeatureReelsResponse `json:"2,omitempty"`
	ThreeWilds *FeatureReelsResponse `json:"3,omitempty"`
}
type FeatureReelsResponse struct {
	Desc string `json:"desc,omitempty"`
	Reel []int  `json:"reel,omitempty"`
}

type PrizeDefResponse struct {
	Multiplier int    `json:"multiplier"`
	Spins      int    `json:"spins"`
	Type       string `json:"type"`
}

func GetEngineDefResponse(engineConf engine.EngineConfig, engineID string) map[string]EngineDefResponse {

	resp := make(map[string]EngineDefResponse)
	switch engineID {
	case "mvgEngineI", "mvgEngineIII", "mvgEngineV", "mvgEngineVII", "mvgEngineIX", "mvgEngineXII", "mvgEngineXIV":
		for i := 0; i < len(engineConf.EngineDefs); i++ {
			reelsetType := engineConf.EngineDefs[i].ID
			if reelsetType == "freespin" {
				reelsetType = "freeSpin"
			}
			resp[fmt.Sprintf("%v", engineConf.EngineDefs[i].Index)] = EngineIDefResponse{engineConf.EngineDefs[i].Reels, reelsetType}
		}
		return resp
	case "mvgEngineII":
		// add first engine def as normal
		resp["0"] = EngineIIDefResponse{Reels: engineConf.EngineDefs[0].Reels, Type: "Base"}
		var featureWildDefs EngineIIDefResponse
		featureWildDefs.NoWilds = &FeatureReelsResponse{"0_wilds", engineConf.EngineDefs[1].Reels[0]}
		featureWildDefs.OneWild = &FeatureReelsResponse{"1_wilds", engineConf.EngineDefs[1].Reels[1]}
		featureWildDefs.TwoWilds = &FeatureReelsResponse{"2_wilds", engineConf.EngineDefs[1].Reels[2]}
		featureWildDefs.ThreeWilds = &FeatureReelsResponse{"3_wilds", engineConf.EngineDefs[1].Reels[3]}

		resp["FeatureWildReelDef"] = featureWildDefs
		return resp
	case "mvgEngineX":
		// add first engine def as normal
		resp["0"] = EngineIIDefResponse{Reels: engineConf.EngineDefs[0].Reels, Type: "base"}
		resp["1"] = EngineIIDefResponse{Reels: engineConf.EngineDefs[1].Reels, Type: "base"}
		return resp
	case "mvgEngineXIII":
		for i := 0; i < len(engineConf.EngineDefs); i++ {
			reelsetType := engineConf.EngineDefs[i].ID
			if reelsetType == "freespin" {
				reelsetType = "freeSpin"
			} else if reelsetType == "base" {
				reelsetType = "Base"
			}
			resp[fmt.Sprintf("%v", engineConf.EngineDefs[i].Index)] = EngineIDefResponse{engineConf.EngineDefs[i].Reels, reelsetType}
		}
		return resp
	}

	return resp

}

func GetPrizeDefResponse(engineConf engine.EngineConfig, engineID string) map[string]PrizeDefResponse {
	logger.Debugf("GetPrizeDefResponse ID: %v", engineID)

	resp := make(map[string]PrizeDefResponse)
	switch engineID {
	case "mvgEngineIII":
		for i := 0; i < len(engineConf.EngineDefs); i++ {
			if engineConf.EngineDefs[i].ID == "pickSpins" {
				for x := 0; x < len(engineConf.EngineDefs[i].SpecialPayouts); x++ {
					multiplier := engineConf.EngineDefs[i].SpecialPayouts[x].Multiplier
					prizepicks := strings.Split(engineConf.EngineDefs[i].SpecialPayouts[x].Index, ":")
					spins, _ := strconv.Atoi(prizepicks[1])
					resp[fmt.Sprintf("freeSpins%s", prizepicks[1])] = PrizeDefResponse{Type: "pickType", Multiplier: multiplier, Spins: spins}
				}
			}
		}
		return resp
	}
	return resp
}

// ParameterResponse ..
type ParameterResponse struct {
	StakeValues  string      `json:"stakeValues"`
	DefaultStake float32     `json:"defaultStake"`
	SessionID    store.Token `json:"host/verified-token"`
}

// PlayerResponse ..
type PlayerResponse struct {
	ID         string          `json:"id"`
	Balance    BalanceResponse `json:"balance"`
	Level      LevelResponse   `json:"level"`
}

type LevelResponse struct {
	Level          int32 `json:"level"`
	Stage          int32 `json:"stage"`
	MaxLevel       int32 `json:"max_level"`
	RemainingSpins int32 `json:"remaining_spins"`
	SpinsToStageUp int32 `json:"spins_to_stageup"`
	TotalSpins     int32 `json:"total_spins"`
}

// BalanceResponse ...
type BalanceResponse struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	FreeGames int    `json:"free_games"`
}

// Render game init response
func (gi GameInitResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GameInfoResponse struct {
	CCY           string `json:"currency"`
	HostName      string `json:"host"`
	InterfaceName string `json:"interface"`
	GameName      string `json:"name"`
	Version       string `json:"version"`
}

// GameplayResponse response struct
type GameplayResponse struct {
	Game   GameInfoResponse `json:"game"`
	Schema string           `json:"$schema,omitempty"`

	// Gamestates []*Gamestate `json:"gamestates"`
	//Gameplay *Gameplay      `json:"gameplay"`
	Player PlayerResponse `json:"player"`
	Status string         `json:"status"`
	//Balance  float32   `json:"balance"`
	GamestateInfo GamestateResponse `json:"gamestate"`
	Parameters    ParameterResponse `json:"parameters,omitempty"`
	Links         []LinkResponse    `json:"_links"`
	Forms         FormResponse      `json:"_forms"`
}

// GamestateResponse ...
type GamestateResponse struct {
	Id              string  `json:"id"`
	Action          string  `json:"action"`
	CurrentPlay     int     `json:"currentPlay"`
	CurrentWinnings float32 `json:"currentWinnings"`
	//freespinWinnings 0
	NumFreeSpins     int     `json:"numFreeSpins"`
	FreeSpinWinnings float32 `json:"freespinWinnings"`
	Stake            float32 `json:"stake"`
	StakePerLine     float32 `json:"stakePerLine"`
	TotalStake       float32 `json:"totalStake"`
	//stopList: [23, 0, 55, 16, 8]
	//totalStake 3
	TotalWinnings        float32       `json:"totalWinnings"` // 4
	View                 [][]string    `json:"view"`          //[["4", "4", "3", "0", "6"], ["7", "8", "6", "3", "5"], ["3", "3", "2", "6", "0"]]
	WildSingleMultiplier int           `json:"wildSingleMultiplier"`
	Wins                 []WinResponse `json:"wins"`
	StopList             []int         `json:"stopList"`
	ReelSetIndex         int           `json:"reelSetIndex"`
	SelectedWinLines     []string      `json:"selectedWinLines"`
	FreeSpinMultiplier   int           `json:"freespin_multiplier,omitempty"`
	Win                  float32       `json:"win"`
	//parameters: {}
	FreeSpinReelset *WinRSResponse `json:"freespinReelset,omitempty"`
}
type WinRSResponse struct {
	Index int     `json:"index"`
	Reels [][]int `json:"reels"`
	Type  string  `json:"type"`
}

// PriceResponse
type PrizeResponse struct {
	ID         string `json:"id"`
	Multiplier int    `json:"multiplier"`
	NumSpins   int    `json:"numSpins"`
	Type       string `json:"type"`
}

// WinResponse ...
type WinResponse struct {
	ID              string         `json:"id"`           // "3:3"
	PayoutFactor    int            `json:"payoutFactor"` // 10
	PrizeResponse   *PrizeResponse `json:"prize,omitempty"`
	ScatterWinnings float32        `json:"ScatterWinnings"`
	Stake           float32        `json:"stake"`
	Type            string         `json:"type"`
	Symbol          int            `json:"symbol"`
	SymbolPos       []int          `json:"symbolPos"`      // ["0", "1", "12"]
	WildMultiplier  int            `json:"wildMultiplier"` // 1
	Winnings        float32        `json:"winnings"`       // 0.1
	WinLine         *int           `json:"winLine,omitempty"`
	StakeIndex      string         `json:"stake_index,omitempty"`
	Frequency       string         `json:"freq, omitempty"`
	//Type string `json:"type"`
}

func fillGamestateResponse(engineConf engine.EngineConfig, gamestate engine.Gamestate) GamestateResponse {

	view := make([][]string, len(gamestate.SymbolGrid[0]))
	for _, row := range gamestate.SymbolGrid {
		for j, symbol := range row {
			view[j] = append(view[j], strconv.Itoa(int(symbol)))
		}
	}
	numFs := 0
	for _, action := range gamestate.NextActions {
		if strings.Contains(action, "freespin") {
			numFs++
		}
	}
	gameSlug, rsID := engine.GetGameIDAndReelset(gamestate.GameID)

	var winType string
	if strings.Contains(engineConf.EngineDefs[0].WinType, "ways") {
		winType = "way"
	} else {
		winType = "line"
	}

	stakeDivisor := engine.NewFixedFromInt(engineConf.EngineDefs[0].StakeDivisor)
	if len(gamestate.SelectedWinLines) > 0 {
		stakeDivisor = engine.NewFixedFromInt(len(gamestate.SelectedWinLines))
	}

	stake := gamestate.BetPerLine.Amount.Mul(stakeDivisor)
	wins := make([]WinResponse, len(gamestate.Prizes))
	for i, p := range gamestate.Prizes {
		p.Index = engineConf.DetectSpecialWins(rsID, p)

		adjustedSymbolPositions := make([]int, len(p.SymbolPositions))

		if len(engineConf.EngineDefs[rsID].ViewSize) == 3 {
			if strings.Contains(gameSlug, "seasons") {
				switch p.Winline {
				case 0:
					adjustedSymbolPositions = []int{3, 4, 5}
				case 1:
					adjustedSymbolPositions = []int{0, 1, 2}
				case 2:
					adjustedSymbolPositions = []int{6, 7, 8}
				}
			} else {
				// for engine IX
				adjustedSymbolPositions = []int{0, 1, 2}
			}

		} else if engineConf.EngineDefs[rsID].ViewSize[0] == 4 {
			// [0,1,2,3,4][5,6,7,8,9][10,11,12,13,14][15,16,17,18,19]
			// [0,4,8,12,16][1,5,9,13,17][2,6,10,14,18][3,7,11,15,19]
			for i, pos := range p.SymbolPositions {

				switch pos {
				case 0, 4, 16, 8, 12:
					adjustedSymbolPositions[i] = pos / 4
				case 1, 5, 9, 13, 17:
					adjustedSymbolPositions[i] = (pos-1)/4 + 5
				case 2, 6, 10, 14, 18:
					adjustedSymbolPositions[i] = (pos-2)/4 + 10
				case 3, 7, 11, 15, 19:
					adjustedSymbolPositions[i] = (pos-3)/4 + 15
				}
			}
		} else {
			// adjust symbol positions for old grid system
			// [0,1,2,3,4][5,6,7,8,9][10,11,12,13,14]
			// [0,3,6,9,12][1,4,7,10,13][2,5,8,11,14]
			for i, pos := range p.SymbolPositions {
				switch pos {
				case 0, 3, 6, 9, 12:
					adjustedSymbolPositions[i] = pos / 3
				case 1, 4, 7, 10, 13:
					adjustedSymbolPositions[i] = (pos-1)/3 + 5
				case 2, 5, 8, 11, 14:
					adjustedSymbolPositions[i] = (pos-2)/3 + 10
				}

			}
		}

		winnings := engine.NewFixedFromInt(p.Payout.Multiplier * p.Multiplier * gamestate.Multiplier).Mul(gamestate.BetPerLine.Amount)

		win := WinResponse{
			ID:              p.Index,
			Symbol:          p.Payout.Symbol,
			Type:            winType,
			PayoutFactor:    p.Payout.Multiplier,
			ScatterWinnings: 0,
			Stake:           stake.ValueAsFloat(),
			SymbolPos:       adjustedSymbolPositions,
			WildMultiplier:  p.Multiplier,
			Winnings:        winnings.ValueAsFloat(),
		}

		if strings.Contains(p.Index, "freespin") {
			win.Type = "scatter"
			w := winnings.Mul(stakeDivisor).ValueAsFloat()
			win.Winnings = w
			win.ScatterWinnings = w

			prizeID := strings.Split(p.Index, ":")
			nSpin, _ := strconv.Atoi(prizeID[1])
			pr := PrizeResponse{
				ID:         fmt.Sprintf("FreeSpins%s", prizeID[1]),
				Multiplier: p.Multiplier,
				NumSpins:   nSpin,
				Type:       "freespins",
			}

			win.PrizeResponse = &pr

		} else if strings.Contains(p.Index, "pickSpins") {
			win.Type = "scatter"
			w := winnings.Mul(stakeDivisor).ValueAsFloat()
			win.Winnings = w
			win.ScatterWinnings = w
			pr := PrizeResponse{
				ID:         "PickSpins",
				Multiplier: 0,
				NumSpins:   0,
				Type:       "featurepick",
			}
			win.PrizeResponse = &pr
		}
		if win.Type == "line" {
			wl := p.Winline
			win.WinLine = &wl
		}
		if strings.Contains(gameSlug, "seasons") {
			win.Stake = gamestate.BetPerLine.Amount.ValueAsFloat()
		}
		wins[i] = win

	}

	// get reel set index
	engineID, err := config.GetEngineFromGame(gameSlug)

	//legacy adapters, in future client should recognize "base" and stake=0 on bonus rounds
	action := "spin"
	if gamestate.Action != "base" && gamestate.Action != "maxBase" {
		if engineID == "mvgEngineIII" {
			switch gamestate.Action {
			case "pickSpins":
				action = "feature_select"
			default:
				action = "freespin"
			}
		} else {
			action = gamestate.Action
		}
	}

	currentWinnings, currentStake := engine.GetCurrentWinAndStake(gamestate)
	if gamestate.Action != "base" {
		currentStake = gamestate.BetPerLine.Amount.Mul(engine.NewFixedFromInt(engineConf.EngineDefs[0].StakeDivisor))
	}

	totalWinnings := gamestate.CumulativeWin
	selectedWinLines := make([]string, 0)
	for _, el := range gamestate.SelectedWinLines {
		selectedWinLines = append(selectedWinLines, strconv.Itoa(el))
	}

	gsResponse := GamestateResponse{
		Id:                   gamestate.Id,
		Action:               action,
		CurrentPlay:          gamestate.PlaySequence,
		CurrentWinnings:      currentWinnings.ValueAsFloat(),
		NumFreeSpins:         numFs,
		FreeSpinWinnings:     0.00,
		View:                 view,
		Stake:                currentStake.ValueAsFloat(),
		StakePerLine:         gamestate.BetPerLine.Amount.ValueAsFloat(),
		TotalStake:           stake.ValueAsFloat(),
		WildSingleMultiplier: gamestate.Multiplier,
		Wins:                 wins,
		TotalWinnings:        totalWinnings.ValueAsFloat(),
		StopList:             gamestate.StopList,
		ReelSetIndex:         rsID,
		SelectedWinLines:     selectedWinLines,
		Win:                  currentWinnings.ValueAsFloat(),
	}
	if engineID == "mvgEngineX" {
		//another hack for the old client
		gsResponse.ReelSetIndex = 0
	}

	if strings.HasPrefix(action, "freespin") {
		// hack for old client support, in future just return round multiplier
		gsResponse.FreeSpinMultiplier = gamestate.Multiplier

		if err != nil {
			logger.Errorf("error retrieving engine id")
			return gsResponse
		}
		if engineID == "mvgEngineII" {
			wildProfile := getWildProfile(gamestate.SymbolGrid)
			var reels [][]int
			for i := 0; i < len(wildProfile); i++ {
				reels = append(reels, engineConf.EngineDefs[1].Reels[wildProfile[i]])
			}
			gsResponse.FreeSpinReelset = &WinRSResponse{
				Index: 1,
				Reels: reels,
				Type:  "Freespins",
			}
		}
	}

	return gsResponse
}

func getWildProfile(symbolGrid [][]int) []int {
	var wildProfile []int
	for i := 0; i < len(symbolGrid); i++ {
		var numWilds int
		for j := 0; j < len(symbolGrid[i]); j++ {
			if symbolGrid[i][j] == 10 {
				numWilds++
			}

		}
		wildProfile = append(wildProfile, numWilds)
	}
	return wildProfile
}

// Render Gameplay Resonse
func (gp GameplayResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// StateResponse response struct
type StateResponse struct {
	State []byte `json:"state"`
}

// Render StateResponse Response
func (st StateResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GameplaySizeResponse response struct
type GamestateSizeResponse struct {
	Size  int `json:"size"`
	Size2 int `json:"size2"`
}

// LinkResponse Link reponse struct
type LinkResponse struct {
	Href   string `json:"href"`
	Method string `json:"method"`
	Rel    string `json:"rel"`
	ID     string `json:"id"`
	Type   string `json:"type"`

}

//FormResponse ...
type FormResponse struct {
	AcceptHeader  string                     `json:"application/octet-stream"`
	FreeSpin      FreeSpinFormResponse       `json:"application/vnd.maverick.slots.freespin-v1+json"`
	NormalSpin    NormalSpinFormResponse     `json:"application/vnd.maverick.slots.spin-v1+json"`
	FeatureSelect *FeatureSelectFormResponse `json:"application/vnd.maverick.slots.feature-select-v1+json,omitempty"`
}

// SpinFormResponse ...
type NormalSpinFormResponse struct {
	Schema           string  `json:"$schema,omitempty"`
	Action           string  `json:"action,omitempty"`
	PerLine          bool    `json:"perLine,omitempty"`
	SelectedWinLines []int   `json:"selectedWinLines,omitempty"`
	Stake            *string `json:"stake"`
}

// FreeSpinFormResponse ...
type FreeSpinFormResponse struct {
	Schema string `json:"$schema,omitempty"`
	Action string `json:"action,omitempty"`
}

type FeatureSelectFormResponse struct {
	Schema          string  `json:"$schema"`
	Action          string  `json:"action"`
	SelectedFeature *string `json:"selectedFeature"`
}

// Render Get Game response
func (gl GameLinkResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render GamestateSize response
func (gps GamestateSizeResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// EngineConfigResponse response struct
type EngineConfigResponse struct {
	EngineConfig *engine.EngineConfig `json:"configs"`
}

// Render EngineConfig response
func (ec EngineConfigResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render Gamestate response
func (ec GamestateResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func BuildForm(g engine.Gamestate, engineID string, showFeature ...bool) FormResponse {

	show := false
	if len(showFeature) > 0 {
		show = showFeature[0]
	}
	forms := FormResponse{}

	// octet-stream
	forms.AcceptHeader = ""
	forms.FreeSpin = FreeSpinFormResponse{Schema: "http://json-schemas.maverick.com/slots/freespin-v1.json", Action: "freespin"}

	forms.NormalSpin = NormalSpinFormResponse{Schema: "http://json-schemas.maverick.com/slots/spin-v1.json", Action: "spin", SelectedWinLines: make([]int, 0), PerLine: true, Stake: nil}

	if engineID == "mvgEngineIII" && show {
		forms.FeatureSelect = &FeatureSelectFormResponse{Schema: "http://json-schemas.maverick.io/slots/feature-select-v1.json", Action: "feature_select", SelectedFeature: nil}
	}

	return forms
}

func renderGamestate(request *http.Request, gamestate engine.Gamestate, balance BalanceResponse, engineConf engine.EngineConfig, playerStore store.PlayerStore) GameplayResponse {
	// generate gamestate response
	operator := request.FormValue("operator")
	mode := chi.URLParam(request, "wallet")
	authID := playerStore.Token
	gameID := strings.Split(gamestate.GameID, ":")[0]
	engineID, _ := config.GetEngineFromGame(gameID)
	urlScheme := GetURLScheme(request)
	status := "FINISHED"

	if gamestate.NextActions[0] != "finish" {
		status = "OPEN"
	} else if gamestate.Closed == true {
		status = "CLOSED"
	}

	level, stage := gamestate.Gamification.GetLevelAndStage()
	remainingSpins := gamestate.Gamification.GetSpins()
	stageUpSpins := gamestate.Gamification.GetSpinsToStageUp()
	totalSpins := gamestate.Gamification.GetTotalSpins()

	player := PlayerResponse{
		ID:         playerStore.PlayerId,
		Balance:    balance,
		Level: LevelResponse{
			Level:          level,
			Stage:          stage,
			MaxLevel:       1000000,
			RemainingSpins: remainingSpins,
			SpinsToStageUp: stageUpSpins,
			TotalSpins:     totalSpins,
		},
	}
	playHref := fmt.Sprintf("%s%s/%s/rgs/play/%s/%s/%s", urlScheme, request.Host, APIVersion, gameID, gamestate.Id, mode)
	// determine if this is the first sham gamestate being rendered:
	if len(gamestate.Transactions) == 0 {
		playHref += fmt.Sprintf("?playerId=%v&ccy=%v&betLimitCode=%v", player.ID, player.Balance.Currency, playerStore.BetLimitSettingCode)
		if playerStore.FreeGames.NoOfFreeSpins > 0 {
			playHref +=fmt.Sprintf("&campaign=%v&numFG=%v", playerStore.FreeGames.CampaignRef, playerStore.FreeGames.NoOfFreeSpins)
		}
		logger.Debugf("Rendering sham init gamestate: %v", gamestate)

	}
	gameInfo := &GameInfoResponse{CCY: player.Balance.Currency, HostName: operator, InterfaceName: mode, GameName: gameID, Version: "2"}
	gpResponse := GameplayResponse{
		Game:          *gameInfo,
		Schema:        DefaultSchema,
		Player:        player,
		GamestateInfo: fillGamestateResponse(engineConf, gamestate),
		Status:        status,
		Parameters:    ParameterResponse{SessionID: authID},
		Links: []LinkResponse{
			{
				Href: fmt.Sprintf("%s%s/%s/gameplay/%v", urlScheme, request.Host, APIVersion, gamestate.Id),
				Rel:  "self",
			},
			{
				Href:   playHref,
				Method: "POST",
				Rel:    "new-game",
				Type:   "application/vnd.maverick.slots.spin-v1+json",
			}, {
				Href:   fmt.Sprintf("%s%s/%s/rgs/clientstate/%s/%s/%s/%s", urlScheme, request.Host, APIVersion, gamestate.Id, authID, gameID, mode),
				Method: "PUT",
				Rel:    "gameplay-client-state-save",
				Type:   "application/octet-stream",
			},
		},
	}
	// add freespin link and form if applicable

	if len(gamestate.NextActions) > 1 {
		// hack to display balance as pre-spin amount
		gpResponse.Player.Balance.Amount = removeCumulativeWinFromBalance(gpResponse.Player.Balance.Amount, gpResponse.GamestateInfo.TotalWinnings)

		gpResponse.Links[1].Rel = "option"
		if strings.Contains(gamestate.NextActions[0], "freespin") ||  strings.Contains(gamestate.NextActions[0], "cascade"){
			gpResponse.Links[1].Type = "application/vnd.maverick.slots.freespin-v1+json"
		} else if strings.Contains(gamestate.NextActions[0], "pick") {
			gpResponse.Links[1].Type = "application/vnd.maverick.slots.feature-select-v1+json"
		}
	}

	forms := BuildForm(gamestate, engineID, true)
	gpResponse.Forms = forms
	return gpResponse
}


func removeCumulativeWinFromBalance(balance string, cumulativeWin float32) string {
	// for legacy client, balance returned on freespin should be initial balance on round commencement
	balanceFloat, err := strconv.ParseFloat(balance, 32)
	if err != nil {
		logger.Errorf("Error converting balance to pre-freegame value: %v", err)
		return balance
	}
	logger.Debugf("converted initial balance %v to final balance %v (diff of cumulativewin %v", balance, fmt.Sprintf("%.2f", float32(balanceFloat)-cumulativeWin), cumulativeWin)
	return fmt.Sprintf("%.2f", float32(balanceFloat)-cumulativeWin)
}