package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_INSTA_WIN = "InstaWin"

	PARAM_ID_INSTA_WIN_AMOUNT    = "InstaWinAmount"
	PARAM_ID_INSTA_WIN_PAYOUTS   = "Payouts"
	PARAM_ID_INSTA_WIN_TYPE      = "InstaWinType"
	PARAM_ID_INSTA_WIN_SOURCE_ID = "InstaWinSourceId"
	PARAM_ID_INSTA_WIN_TILE_ID   = "TileId"
	PARAM_ID_INSTA_WIN_POSITIONS = "Positions"
	PARAM_ID_INSTA_WIN_INDEX     = "PayoutIndex"

	PARAM_VALUE_INSTA_WIN_BONUS = "bonus"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_INSTA_WIN, func() feature.Feature { return new(InstaWin) })

type InstaWinData struct {
	Type     string `json:"type"`
	SourceId int32  `json:"sourceid"`
	Amount   int    `json:"amount"`
	Payouts  []int  `json:"payouts,omitempty"`
}

type InstaWin struct {
	feature.Base
	Data InstaWinData `json:"data"`
}

func (f *InstaWin) DataPtr() interface{} {
	return &f.Data
}

func (f InstaWin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	multiplier := params.GetInt(PARAM_ID_INSTA_WIN_AMOUNT)
	var payouts []int
	var index string
	if params.HasKey(PARAM_ID_INSTA_WIN_PAYOUTS) {
		payouts = params.GetIntSlice(PARAM_ID_INSTA_WIN_PAYOUTS)
	}
	if params.HasKey(PARAM_ID_INSTA_WIN_INDEX) {
		index = params.GetString(PARAM_ID_INSTA_WIN_INDEX)
	}
	state.Features = append(state.Features,
		&InstaWin{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: InstaWinData{
				Type:     params.GetString(PARAM_ID_INSTA_WIN_TYPE),
				SourceId: params.GetInt32(PARAM_ID_INSTA_WIN_SOURCE_ID),
				Amount:   multiplier,
				Payouts:  payouts,
			},
		})
	state.Wins = append(state.Wins,
		feature.FeatureWin{
			Index:           index,
			Multiplier:      multiplier,
			Symbols:         []int{params.GetInt(PARAM_ID_INSTA_WIN_TILE_ID)},
			SymbolPositions: params.GetIntSlice(PARAM_ID_INSTA_WIN_POSITIONS),
		})
}

func (f *InstaWin) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *InstaWin) Deserialize(data []byte) (err error) {
	return feature.DeserializeFeatureFromBytes(f, data)
}
