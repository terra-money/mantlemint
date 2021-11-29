package rpc

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"net/http"
	"net/http/httptest"
	"sync"
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
	mtx             *sync.Mutex
	inTransit       map[string]bool
	resultChan      map[string]chan *ResponseCache
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
		mtx:             new(sync.Mutex),
		inTransit:       make(map[string]bool),
		resultChan:      make(map[string]chan *ResponseCache),
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
	cb.inTransit = make(map[string]bool)
	cb.resultChan = make(map[string]chan *ResponseCache)
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

	// 캐시가 없다 --> 첫번째 리퀘스트이거나 아직 캐시 up이 안된것
	// inTransit = false인 경우 첫 리퀘스트
	// 아니면 중복

	_, isInTransit := cb.inTransit[uri]

	// if isInTransit is false, this is the first time we're processing this query
	// run actual querier
	if !isInTransit {
		// 채널 만듬
		cb.mtx.Lock()
		cb.inTransit[uri] = true
		cb.resultChan[uri] = make(chan *ResponseCache)
		cb.mtx.Unlock()

		recorder := httptest.NewRecorder()

		// process request
		handler.ServeHTTP(recorder, request)

		// set in cache
		responseCacheBody := cb.Set(request.URL.String(), recorder.Code, recorder.Body.Bytes())

		cb.resultChan[uri] <- responseCacheBody

		// write
		writer.WriteHeader(recorder.Code)
		writer.Write(recorder.Body.Bytes())

		cb.mtx.Lock()
		delete(cb.inTransit, uri)
		cb.mtx.Unlock()

		return
	}

	// wait for result
	response := <-cb.resultChan[uri]

	writer.WriteHeader(response.status)
	writer.Write(response.body)
	return
}
