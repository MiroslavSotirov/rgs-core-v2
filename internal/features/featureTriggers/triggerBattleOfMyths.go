package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
)

const (
	FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS = "TriggerBattleOfMyths"

	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FEATURE_PROBABILITY = "FeatureProbability"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_PRINCESS        = "RunPrincess"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_TIGER           = "RunTiger"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_DRAGON          = "RunDragon"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS, func() feature.Feature { return new(TriggerBattleOfMyths) })

type TriggerBattleOfMyths struct {
	feature.Base
}

func (f TriggerBattleOfMyths) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	if len(state.CalculateWins(state.SourceGrid, nil)) > 0 {
		params["PureWins"] = true
	}

	runPrincess := false
	for _, f := range state.Features {
		if f.DefPtr().Type == featureProducts.FEATURE_ID_PRINCESS {
			runPrincess = true
		}
	}

	params[PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_PRINCESS] = runPrincess
	if !runPrincess {
		featureProb := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FEATURE_PROBABILITY)
		random := rng.RandFromRangePool(10000)
		if random < featureProb {
			random = rng.RandFromRangePool(2)
			if random == 0 {
				params[PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_TIGER] = true
			} else {
				params[PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_DRAGON] = true
			}
		}
	}

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerBattleOfMyths) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMyths) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMyths) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
