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
	mtx             *sync.Mutex
	cacheMtxMap     map[string]*sync.Mutex
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
		cacheMtxMap:     make(map[string]*sync.Mutex),
	}
}

func (cb *CacheBackend) Set(cacheKey string, status int, body []byte) {
	if evicted := cb.lru.Add(cacheKey, &ResponseCache{
		status: status,
		body:   body,
	}); evicted != false {
		cb.evictionCount++
	}
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
	cb.cacheMtxMap = make(map[string]*sync.Mutex)
	cb.mtx.Unlock()
}

func (cb *CacheBackend) HandleCachedHTTP(writer http.ResponseWriter, request *http.Request, handler http.Handler) {
	cb.mtx.Lock()
	cb.serveCount++
	cb.mtx.Unlock()

	// set response type as json
	writer.Header().Set("Content-Type", "application/json")

	uri := request.URL.String()

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

	// critical section for checking if caching is in transit
	// and set it true if not
	cb.mtx.Lock()
	mtxForUri, isInTransit := cb.cacheMtxMap[uri]

	// if isInTransit is false, this is the first time we're processing this query
	// run actual querier
	if !isInTransit {
		mtxForUri := &sync.Mutex{}
		cb.cacheMtxMap[uri] = mtxForUri
		mtxForUri.Lock()
		cb.mtx.Unlock()

		recorder := httptest.NewRecorder()

		// process request
		handler.ServeHTTP(recorder, request)

		// set in cache
		cb.Set(request.URL.String(), recorder.Code, recorder.Body.Bytes())

		// write
		writer.WriteHeader(recorder.Code)
		writer.Write(recorder.Body.Bytes())

		mtxForUri.Unlock()
	} else {
		// same query is processing but not cached yet.
		cb.mtx.Unlock()

		// wait for the cache
		mtxForUri.Lock()
		cached := cb.Get(request.URL.String())
		mtxForUri.Unlock()

		if cached == nil {
			panic("cache not set")
		}

		writer.WriteHeader(cached.status)
		writer.Write(cached.body)

		cb.mtx.Lock()
		cb.cacheServeCount++
		cb.mtx.Unlock()

	}

}
