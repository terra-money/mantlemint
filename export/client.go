package export

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	terra "github.com/terra-money/core/v2/app"
)

func ShouldCacheRoute(route string) bool {
	r := strings.Split(route, "/")
	if len(r) > 2 && r[1] == "export" {
		return false
	}
	return true
}

func RegisterRESTRoutes(router *mux.Router, app *terra.TerraApp) {
	router.Handle("/export/accounts", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		err := ExportAllAccounts(app)
		if err != nil {
			writer.WriteHeader(http.StatusConflict)
			writer.Write([]byte(err.Error()))
		}
		writer.WriteHeader(http.StatusOK)
	})).Methods("POST")

	router.Handle("/export/circulating_supply", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cs, err := ExportCirculatingSupply(app)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cs.String()))
	}))
}
