package rpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	lru "github.com/hashicorp/golang-lru"
)

type ResponseCache struct {
	status int
	body   []byte
}

type CacheBackend struct {
	lru             *lru.Cache
	evictionCount   uint64
	cacheServeCount uint64
	serveCount      uint64
	cacheType       string
	mtx             *sync.RWMutex

	// subscribe to cache for same request URI
	resultChan     map[string]chan *ResponseCache
	subscribeCount map[string]int
}

func NewCacheBackend(cacheSize int, cacheType string) *CacheBackend {
	// lru.New
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}

	return &CacheBackend{
		lru:             cache,
		evictionCount:   0,
		cacheServeCount: 0,
		serveCount:      0,
		cacheType:       cacheType,
		mtx:             new(sync.RWMutex),
		resultChan:      make(map[string]chan *ResponseCache),
		subscribeCount:  make(map[string]int),
	}
}

func (cb *CacheBackend) Set(cacheKey string, status int, body []byte) *ResponseCache {
	response := &ResponseCache{
		status: status,
		body:   body,
	}
	if evicted := cb.lru.Add(cacheKey, response); evicted != false {
		cb.evictionCount++
	}

	return response
}

func (cb *CacheBackend) Get(cacheKey string) *ResponseCache {
	cached, ok := cb.lru.Get(cacheKey)
	if !ok {
		return nil
	}

	data, _ := cached.(*ResponseCache)
	return data
}

func (cb *CacheBackend) Metric() {
	fmt.Printf("[rpc/%s] cache length %d, eviction count %d, serveCount %d, cacheServeCount %d\n",
		cb.cacheType,
		cb.lru.Len(),
		cb.evictionCount,
		cb.serveCount,
		cb.cacheServeCount,
	)
}

func (cb *CacheBackend) Purge() {
	cb.mtx.Lock()
	cb.lru.Purge()
	cb.evictionCount = 0
	cb.cacheServeCount = 0
	cb.serveCount = 0
	cb.mtx.Unlock()
}

func (cb *CacheBackend) HandleCachedHTTP(writer http.ResponseWriter, request *http.Request, handler http.Handler) {
	cb.mtx.Lock()
	cb.serveCount++
	cb.mtx.Unlock()

	uri := request.URL.String()

	// see if this request is already made, and in transit
	// set response type as json
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Connection", "close")

	cached := cb.Get(request.URL.String())
	// if cached, return as is
	if cached != nil {
		writer.WriteHeader(cached.status)
		writer.Write(cached.body)

		cb.mtx.Lock()
		cb.cacheServeCount++
		cb.mtx.Unlock()
		return
	}

	cb.mtx.Lock()
	resChan, isInTransit := cb.resultChan[uri]

	// if isInTransit is false, this is the first time we're processing this query
	// run actual querier
	if !isInTransit {
		c := make(chan *ResponseCache)
		cb.resultChan[uri] = c
		cb.subscribeCount[uri] = 0
		cb.mtx.Unlock()

		recorder := httptest.NewRecorder()
		var cache *ResponseCache

		go func() {
			<-request.Context().Done()

			// feed all subscriptions
			cb.mtx.RLock()
			subscribeCount := cb.subscribeCount[uri]
			cb.mtx.RUnlock()

			if cache != nil {
				for i := 0; i < subscribeCount; i++ {
					c <- cache
				}
			}
			close(c)

			cb.mtx.Lock()
			delete(cb.subscribeCount, uri)
			delete(cb.resultChan, uri)
			cb.mtx.Unlock()
		}()

		// process request
		handler.ServeHTTP(recorder, request)

		// set in cache
		cache = cb.Set(request.URL.String(), recorder.Code, recorder.Body.Bytes())

		// write
		writer.WriteHeader(recorder.Code)
		writer.Write(recorder.Body.Bytes())

		return
	}

	// same query is processing but not cached yet.
	// subscribe for cache result here.
	cb.subscribeCount[uri]++
	cb.mtx.Unlock()

	response, ok := <-resChan

	if ok {
		writer.WriteHeader(response.status)
		writer.Write(response.body)
	} else {
		writer.WriteHeader(503)
		writer.Write([]byte("Service Unavailable"))
	}
}
