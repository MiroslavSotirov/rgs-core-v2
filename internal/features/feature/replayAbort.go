package feature

const (
	FEATURE_ID_REPLAY_ABORT = "ReplayAbort"
)

var _ Factory = RegisterFeature(FEATURE_ID_REPLAY_ABORT, func() Feature { return new(ReplayAbort) })

type ReplayAbort struct {
	Base
}

func (f ReplayAbort) Trigger(state *FeatureState, params FeatureParams) {

	state.Replay = false

}

func (f *ReplayAbort) Serialize() ([]byte, error) {
	return SerializeFeatureToBytes(f)
}

func (f *ReplayAbort) Deserialize(data []byte) (err error) {
	return DeserializeFeatureFromBytes(f, data)
}
