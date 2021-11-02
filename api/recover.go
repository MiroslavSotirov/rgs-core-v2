package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

// Recovery extends chi Recoverer middleware with custom response
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			rvr := recover()
			if rvr != nil {

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
					debug.PrintStack()
				}

				rgsInternalServerError := ErrInternalServerError
				rgsInternalServerError.ErrorText = fmt.Sprintf("%s %v", rgsInternalServerError.ErrorText, rvr)
				jsonBody, _ := json.Marshal(ErrInternalServerError)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}

		}()

		next.ServeHTTP(w, r)

	})
}
