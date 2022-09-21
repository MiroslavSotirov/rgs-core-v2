package features

import "fmt"

type Respin struct {
	FeatureDef
	Data FeatureParams `json:"data"`
}

func (f *Respin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *Respin) DataPtr() interface{} {
	return &f.Data
}

func (f *Respin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *Respin) OnInit(state *FeatureState) {
}

func (f Respin) Trigger(state *FeatureState, params FeatureParams) {

	action := params.GetString("Action")
	amount := 1
	if params.HasKey("Amount") {
		amount = params.GetInt("Amount")
	}
	state.Wins = append(state.Wins, FeatureWin{
		Index: fmt.Sprintf("%s:%d", action, amount),
	})
	activateFeatures(f.FeatureDef, state, params)
}

func (f *Respin) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *Respin) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
