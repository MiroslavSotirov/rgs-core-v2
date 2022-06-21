package features

type TriggerBattleOfMythsPrincess struct {
	FeatureDef
}

func (f *TriggerBattleOfMythsPrincess) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMythsPrincess) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMythsPrincess) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMythsPrincess) OnInit(state *FeatureState) {
}

func (f TriggerBattleOfMythsPrincess) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	tileId := params.GetInt("TileId")

	positions := []int{}
	gridh := len(state.SymbolGrid[0])
	for x, r := range state.SymbolGrid {
		fy := -1
		for y, s := range r {
			if fy < 0 {
				if s == tileId {
					fy = y
				}
			} else {
				if s == tileId {
					for ry := fy + 1; ry < y; ry++ {
						positions = append(positions, x*gridh+ry)
					}
					more := false
					for yc := y + 1; !more && yc < len(r); yc++ {
						if r[yc] == tileId {
							more = true
						}
					}
					if len(positions) > 0 {
						params["Positions"] = positions
						params["StartPos"] = x*gridh + fy
						params["EndPos"] = x*gridh + y
						activateFeatures(f.FeatureDef, state, params)
					}
					if more {
						fy = y
					} else {
						break
					}
				}
			}
		}
	}
	gridw := len(state.SymbolGrid)
	for y := 0; y < gridh; y++ {
		fx := -1
		for x := 0; x < gridw; x++ {
			s := state.SymbolGrid[x][y]
			if fx < 0 {
				if s == tileId {
					fx = x
				}
			} else {
				if s == tileId {
					for rx := fx + 1; rx < x; rx++ {
						positions = append(positions, rx*gridh+y)
					}
					more := false
					for xc := x + 1; !more && xc < gridw; xc++ {
						if state.SymbolGrid[xc][y] == tileId {
							more = true
						}
					}
					if len(positions) > 0 {
						params["Positions"] = positions
						params["StartPos"] = fx*gridh + y
						params["EndPos"] = x*gridh + y
						activateFeatures(f.FeatureDef, state, params)
					}
					if more {
						fx = x
					} else {
						break
					}
				}
			}
		}
	}
	return
}

func (f TriggerBattleOfMythsPrincess) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMythsPrincess) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsPrincess) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
