package api

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"net/http"
)

type BalanceCheckResponse struct {
	Token     store.Token  `json:"host/verified-token,omitempty"`
	BalanceResponseV2
}

// TODO: move all common render codes here

// Render Generic Balance Response
func (gb BalanceCheckResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
