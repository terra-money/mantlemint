package block

import (
	"bytes"
	"fmt"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/terra-money/mantlemint-provider-v0.34.x/block_feed"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func SidesyncBlock() {
	targets := getTargets(os.Getenv("SIDESYNC_TARGETS"))
	rpc := getRPCEndpoint(os.Getenv("RPC_ENDPOINT"))
	initialHeight := getInitialHeight(os.Getenv("INITIAL_HEIGHT"))

	targetHeight := initialHeight
	for {
		fmt.Printf("[indexer/block/sidesync] syncing block %d..\n", targetHeight)

		resp, err := http.Get(fmt.Sprintf("%v/block?height=%d", rpc, targetHeight))
		if err != nil {
			panic(err)
		}

		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		block, err := block_feed.ExtractBlockFromRPCResponse(response)
		if err != nil {
			panic(err)
		}

		input, err := tmjson.Marshal(block)
		if err != nil {
			panic(err)
		}

		// if targets len becomes 0, we've finished backfilling
		if len(targets) == 0 {
			break
		}

		for _, target := range targets {
			r, err := http.Post(fmt.Sprintf("%v%s", target, EndpointPOSTBlock), "application/json", bytes.NewBuffer(input))
			if err != nil {
				panic(err)
			}

			// 204 means you can stop calling this target; target already has this data in db
			if r.StatusCode == 204 {
				targets = targets
			}

			if r.StatusCode > 400 {
				bz, _ := ioutil.ReadAll(r.Body)
				fmt.Println(string(bz), r.StatusCode)
				panic("mantle responded with non-ok code")
			}
		}

		targetHeight++
	}

}

func getRPCEndpoint(rpcEndpoint string) string {
	if rpcEndpoint == "" {
		panic("rpc endpoint not set; expected http endpoint, port 26657")
	}
	return rpcEndpoint
}

func getTargets(targets string) []string {
	if targets == "" {
		panic("targets not set; expected a comma delimited list of IPs")
	}

	return strings.Split(targets, ",")
}

func getInitialHeight(height string) int {
	if height == "" {
		panic("initial height not set")
	}

	if h, e := strconv.Atoi(height); e != nil {
		panic(e)
	} else {
		return h
	}
}
