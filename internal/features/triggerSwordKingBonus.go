package features

type TriggerSwordKingBonus struct {
	FeatureDef
}

func (f *TriggerSwordKingBonus) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKingBonus) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKingBonus) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKingBonus) OnInit(state *FeatureState) {
}

func (f TriggerSwordKingBonus) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	number := params.GetIntSlice("NumReels")
	numberProbs := params.GetIntSlice("NumProbabilities")
	numIdx := WeightedRandomIndex(numberProbs)
	num := number[numIdx]

	reelProbs := convertIntSlice(params.GetSlice("ReelProbabilities")[numIdx])
	reelsIdx := WeightedRandomIndex(reelProbs)

	//	logger.Debugf("num: %d reelsIdx: %d ReelPositions: %v", num, reelsIdx, params.GetSlice("ReelPositions"))
	reels := convertIntSlice(params.GetSlice("ReelPositions")[numIdx].([]interface{})[reelsIdx])

	gridh := len(state.SymbolGrid[0])

	positions := []int{}
	for i := 0; i < num; i++ {
		x := reels[i]
		for y := 0; y < 4; y++ {
			positions = append(positions, x*gridh+y)
		}
	}

	if params.GetBool("ResetCounter") {
		SetStatefulStakeMap(*state, FeatureParams{
			"counter": 0,
		}, params)
	}

	if len(positions) > 0 {
		params["Positions"] = positions
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerSwordKingBonus) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKingBonus) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKingBonus) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
