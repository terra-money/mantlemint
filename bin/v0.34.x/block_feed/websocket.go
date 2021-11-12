package block_feed

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

var _ BlockFeed = (*WSSubscription)(nil)

type WSSubscription struct {
	wsEndpoints []string
	ws          *websocket.Conn
	c           chan *BlockResult
}

type handshake struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int    `json:"id"`
	Params  params `json:"params"`
}

type params struct {
	Query string `json:"query"`
}

func NewWSSubscription(wsEndpoints []string) (*WSSubscription, error) {
	return &WSSubscription{
		wsEndpoints: wsEndpoints,
		ws:          nil,
	}, nil
}

func (ws *WSSubscription) Subscribe(rpcIndex int) (chan *BlockResult, error) {
	socket, _, err := websocket.DefaultDialer.Dial(ws.wsEndpoints[rpcIndex], nil)

	// return err, handle failures gracefully
	if err != nil {
		return nil, err
	}

	ws.ws = socket

	var request = &handshake{
		JSONRPC: "2.0",
		Method:  "subscribe",
		ID:      0,
		Params: params{
			Query: "tm.event = 'NewBlock'",
		},
	}

	log.Print("Subscribing to tendermint rpc...")

	// should not fail here
	if err := ws.ws.WriteJSON(request); err != nil {
		return nil, err
	}

	// handle initial message
	// by setting c.initialized to true, we prevent message mishandling
	if err := handleInitialHandhake(ws.ws); err != nil {
		return nil, err
	}

	log.Print("Subscription and the first handshake done. Receiving blocks...")

	// create channel
	c := make(chan *BlockResult)
	ws.c = c

	go receiveBlockEvents(ws.ws, c)

	// start receiving blocks
	return c, nil
}

func (ws *WSSubscription) Close() error {
	return ws.ws.Close()
}

// tendermint rpc sends the "subscription ok" for the intiail response
// filter that out by only sending through channel when there is
// "data" field present
func handleInitialHandhake(ws *websocket.Conn) error {
	_, _, err := ws.ReadMessage()

	if err != nil {
		return err
	}

	return nil
}

// TODO: handle errors here
func receiveBlockEvents(ws *websocket.Conn, c chan *BlockResult) {
	defer close(c)
	for {
		_, message, err := ws.ReadMessage()

		// if read message failed,
		// scrap the whole ws thing
		if err != nil {
			closeErr := ws.Close()
			if closeErr != nil {
				log.Print("websocket close failed, but it seems the underlying websocket is already closed")
			}

			// "reconnect" message!
			c <- nil
			return
		}

		var unmarshalErr error

		// check error
		errorMessage := new(struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    string `json:"data"`
			} `json:"error"`
		})

		if unmarshalErr = json.Unmarshal(message, errorMessage); unmarshalErr != nil {
			panic(unmarshalErr)
		}

		// tendermint has sent error message,
		// close ws
		if errorMessage.Error.Code != 0 {
			log.Printf(
				"tendermint RPC error, code=%d, message=%s, data=%s",
				errorMessage.Error.Code,
				errorMessage.Error.Message,
				errorMessage.Error.Data,
			)
			_ = ws.Close()
		}

		if block, blockParseErr := extractBlockFromWSResponse(message); blockParseErr != nil {
			panic(blockParseErr)
		} else {
			c <- block
		}
	}
}
