package features

import (
	"encoding/json"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type PrincessData struct {
	Positions     []int `json:"positions"`
	StartPos      int   `json:"startpos"`
	EndPos        int   `json:"endpos"`
	ReplaceWithId int   `json:"replacewithid"`
}

type Princess struct {
	FeatureDef
	Data PrincessData `json:"data"`
}

func (f *Princess) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *Princess) DataPtr() interface{} {
	return &f.Data
}

func (f *Princess) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *Princess) OnInit(state *FeatureState) {
}

func (f Princess) forceActivateFeature(featurestate *FeatureState) {
}

func (f Princess) Trigger(state *FeatureState, params FeatureParams) {
	replaceid := params.GetInt("ReplaceWithId")
	positions := params.GetIntSlice("Positions")
	gridh := len(state.SymbolGrid[0])
	for _, p := range positions {
		x := p / gridh
		y := p - (x * gridh)
		state.SymbolGrid[x][y] = replaceid
	}
	state.Features = append(state.Features,
		&Princess{
			FeatureDef: *f.DefPtr(),
			Data: PrincessData{
				Positions:     positions,
				StartPos:      params.GetInt("StartPos"),
				EndPos:        params.GetInt("EndPos"),
				ReplaceWithId: params.GetInt("ReplaceWithId"),
			},
		})
}

// remove this as soon as the duplicates problem has been tracked down
func (f Princess) Validate() {
	duplicates := make(map[int]bool)
	for _, p := range f.Data.Positions {
		_, ok := duplicates[p]
		if ok {
			b, _ := json.Marshal(f)
			logger.Debugf("broken Princess: %s", string(b))
			panic("Princess feature validation failed")
		}
		duplicates[p] = true
	}
}

func (f *Princess) Serialize() ([]byte, error) {
	//	f.Validate()
	return serializeFeatureToBytes(f)
}

func (f *Princess) Deserialize(data []byte) (err error) {
	err = deserializeFeatureFromBytes(f, data)
	//	f.Validate()
	return
}
