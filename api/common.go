package api

import (
	"net/http"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
)

type BalanceCheckResponse struct {
	Token store.Token `json:"host/verified-token,omitempty"`
	BalanceResponseV2
}

type SetBalanceParams struct {
	Balance engine.Money `json:"balance"`
}

type PlayCheckExtParams struct {
	Feeds []store.RestTransactiondata `json:"feeds"`
}

// TODO: move all common render codes here

// Render Generic Balance Response
func (gb BalanceCheckResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
