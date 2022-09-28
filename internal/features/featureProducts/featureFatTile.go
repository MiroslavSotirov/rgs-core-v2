package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"

const (
	FEATURE_ID_FAT_TILE = "FatTile"

	PARAM_ID_FAT_TILE_X       = "X"
	PARAM_ID_FAT_TILE_Y       = "Y"
	PARAM_ID_FAT_TILE_W       = "W"
	PARAM_ID_FAT_TILE_H       = "H"
	PARAM_ID_FAT_TILE_TILE_ID = "TileId"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_FAT_TILE, func() feature.Feature { return new(FatTile) })

type FatTileData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	W      int `json:"w"`
	H      int `json:"h"`
	TileId int `json:"tileid"`
}

type FatTile struct {
	feature.Base
	Data FatTileData `json:"data"`
}

func (f *FatTile) DataPtr() interface{} {
	return &f.Data
}

func (f FatTile) forceActivateFeature(featurestate *feature.FeatureState, x int, y int, tileid int) {
	gridw, gridh := len(featurestate.SymbolGrid), len(featurestate.SymbolGrid[0])
	tilew, tileh := f.Data.W, f.Data.H
	for r := x; r < x+tilew && r < gridw; r++ {
		for s := y; s < y+tileh && s < gridh; s++ {
			featurestate.SymbolGrid[r][s] = tileid
		}
	}
}

func (f FatTile) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	x := params.GetInt(PARAM_ID_FAT_TILE_X)
	y := params.GetInt(PARAM_ID_FAT_TILE_Y)
	w := params.GetInt(PARAM_ID_FAT_TILE_W)
	h := params.GetInt(PARAM_ID_FAT_TILE_H)
	tileid := params.GetInt(PARAM_ID_FAT_TILE_TILE_ID)
	gridw := len(state.SymbolGrid)
	for r := max(x, 0); r < min(x+w, gridw); r++ {
		gridh := len(state.SymbolGrid[r])
		for s := max(y, 0); s < min(y+h, gridh); s++ {
			state.SymbolGrid[r][s] = tileid
		}
	}
	state.Features = append(state.Features,
		&FatTile{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: FatTileData{
				X:      x,
				Y:      y,
				W:      w,
				H:      h,
				TileId: tileid,
			},
		})
}

func (f *FatTile) Serialize() ([]byte, error) {
	return feature.SerializeFeatureToBytes(f)
}

func (f *FatTile) Deserialize(data []byte) (err error) {
	return feature.DeserializeFeatureFromBytes(f, data)
}
