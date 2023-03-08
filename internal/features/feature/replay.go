package feature

const (
	FEATURE_ID_REPLAY = "Replay"

	FEATURE_ID_REPLAY_REPLAY_TRIES  = "ReplayTries"
	FEATURE_ID_REPLAY_REPLAY_PARAMS = "ReplayParams"
)

var _ Factory = RegisterFeature(FEATURE_ID_REPLAY, func() Feature { return new(Replay) })

type Replay struct {
	Base
}

func (f Replay) Trigger(state *FeatureState, params FeatureParams) {

	state.Replay = state.ReplayTries <= params.GetInt(FEATURE_ID_REPLAY_REPLAY_TRIES)

	if params.HasKey(FEATURE_ID_REPLAY_REPLAY_PARAMS) {
		state.ReplayParams = params.GetParams(FEATURE_ID_REPLAY_REPLAY_PARAMS)
	}

}

func (f *Replay) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *Replay) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}
