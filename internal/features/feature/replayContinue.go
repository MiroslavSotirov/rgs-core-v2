package feature

const (
	FEATURE_ID_REPLAY_CONTINUE = "ReplayContinue"
)

var _ Factory = RegisterFeature(FEATURE_ID_REPLAY_CONTINUE, func() Feature { return new(ReplayContinue) })

type ReplayContinue struct {
	Base
}

func (f ReplayContinue) Trigger(state *FeatureState, params FeatureParams) {

	state.Replay = true

}

func (f *ReplayContinue) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *ReplayContinue) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}
