package api

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"net/http"
)

type GenericBalanceResponse struct {
	Token    store.Token  `json:"host/verified-token"`
	Amount   engine.Fixed `json:"amount"`
	Currency string       `json:"currency"`
}

// TODO: move all common render codes here

// Render Generic Balance Response
func (gb GenericBalanceResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
