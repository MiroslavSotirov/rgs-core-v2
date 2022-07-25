package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerSwordKingBonusScatter struct {
	FeatureDef
}

func (f *TriggerSwordKingBonusScatter) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKingBonusScatter) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKingBonusScatter) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKingBonusScatter) OnInit(state *FeatureState) {
}

func (f TriggerSwordKingBonusScatter) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	if state.PureWins ||
		(params.HasKey("RunWilds") && params.GetBool("RunWilds")) ||
		(params.HasKey("RunRespin") && params.GetBool("RunRespin")) {
		logger.Debugf("skipping bonus scatters")
		return
	}

	Probability := params.GetInt("Probability")
	if rng.RandFromRange(10000) < Probability {

		activate := false
		positions := []int{}
		if !params.HasKey("NumScatters") {
			activate = true
		} else {
			numScatters := params.GetIntSlice("NumScatters")[WeightedRandomIndex(params.GetIntSlice("NumProbabilities"))]
			gridh := len(state.SymbolGrid[0])

			logger.Debugf("placing %d bonus scatters", numScatters)

			reels := RandomPermutation([]int{0, 1, 2, 3, 4})

			for s := 0; s < numScatters; s++ {
				reel := reels[s]
				row := rng.RandFromRange(4)
				pos := reel*gridh + row
				positions = append(positions, pos)
			}

			if len(positions) > 0 {
				params["RunBonusScatter"] = true
				params["Positions"] = positions
				logger.Debugf("positions: %v", positions)
				activateFeatures(f.FeatureDef, state, params)

				var counter int
				statefulStake := GetStatefulStakeMap(*state)
				if statefulStake.HasKey("counter") {
					counter = statefulStake.GetInt("counter")
				}
				counter += len(positions)

				counterMax := params.GetInt("CounterMax")
				activate = counter >= counterMax
				SetStatefulStakeMap(*state, FeatureParams{
					"counter": counter,
				}, params)
			}
		}

		if activate {
			logger.Debugf("activate bonus")
			params["RunBonus"] = true
			fstype := params.GetString("FSType")
			numFreespins := params.GetInt("NumFreespins")

			state.Wins = append(state.Wins, FeatureWin{
				Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
				SymbolPositions: positions,
			})
		}
	}

	return
}

func (f TriggerSwordKingBonusScatter) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKingBonusScatter) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKingBonusScatter) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
