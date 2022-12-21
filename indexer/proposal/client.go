package proposal

import (
	"fmt"
	"github.com/gorilla/mux"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/indexer"
	"net/http"
	"strconv"
)

var (
	EndpointGETProposal = "/index/proposal/{id}"
)

var (
	ErrorProposalId       = func(id string) string { return fmt.Sprintf("invalid proposal id %s", id) }
	ErrorRichlistNotFound = func(height string) string { return fmt.Sprintf("richlist at %s not found... yet.", height) }
	EmptyVotes            = []Vote{}
)

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, db tmdb.DB) {
	router.HandleFunc(EndpointGETProposal, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		proposalIdStr, ok := vars["id"]
		if !ok {
			http.Error(writer, ErrorProposalId(proposalIdStr), 400)
			return
		}

		proposalId, err := strconv.ParseUint(proposalIdStr, 10, 64)
		if err != nil {
			http.Error(writer, ErrorProposalId(proposalIdStr), 400)
			return
		}
		proposal, err := GetProposal(db, proposalId)
		if err != nil {
			// instead of returning error, we follow api service and return empty array
			res, err := tmjson.Marshal(EmptyVotes)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}
			writer.WriteHeader(200)
			writer.Write(res)
			return
		}

		res, err := tmjson.Marshal(proposal.Votes)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		writer.WriteHeader(200)
		writer.Write(res)
	})
})
