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
	mtx             *sync.Mutex
}

func NewCacheBackend(cacheSize int) *CacheBackend {
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
		mtx:             new(sync.Mutex),
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

func (cb *CacheBackend) Purge() int {
	fmt.Printf("[rpc/cache] cache eviction count %d, serveCount %d, cacheServeCount %d\n",
		cb.evictionCount,
		cb.serveCount,
		cb.cacheServeCount,
	)

	cb.mtx.Lock()
	cacheLen := cb.lru.Len()
	cb.lru.Purge()
	cb.evictionCount = 0
	cb.cacheServeCount = 0
	cb.serveCount = 0
	cb.mtx.Unlock()
	return cacheLen
}

func (cb *CacheBackend) HandleCachedHTTP(writer http.ResponseWriter, request *http.Request, handler http.Handler) {
	cb.mtx.Lock()
	cb.serveCount++
	cb.mtx.Unlock()

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

	recorder := httptest.NewRecorder()

	// process request
	handler.ServeHTTP(recorder, request)

	// set in cache
	cb.Set(request.URL.String(), recorder.Code, recorder.Body.Bytes())

	// write
	writer.WriteHeader(recorder.Code)
	writer.Write(recorder.Body.Bytes())
}
