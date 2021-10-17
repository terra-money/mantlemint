package rpc

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"net/http"
	"net/http/httptest"
)

type ResponseCache struct {
	status int
	body   []byte
}

type CacheBackend struct {
	lru           *lru.Cache
	evictionCount uint64
}

func NewCacheBackend(cacheSize int) *CacheBackend {
	// lru.New
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}

	return &CacheBackend{
		lru:           cache,
		evictionCount: 0,
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
	defer fmt.Printf("[rpc/cache] cache eviction count %d\n", cb.evictionCount)
	cacheLen := cb.lru.Len()
	cb.lru.Purge()
	cb.evictionCount = 0
	return cacheLen
}

func (cb *CacheBackend) HandleCachedHTTP(writer http.ResponseWriter, request *http.Request, handler http.Handler) {
	cached := cb.Get(request.URL.String())
	// if cached, return as is
	if cached != nil {
		writer.WriteHeader(cached.status)
		writer.Write(cached.body)
		return
	}

	recorder := httptest.NewRecorder()

	// process request
	handler.ServeHTTP(recorder, request)

	// set in cache
	cb.Set(request.URL.String(), recorder.Code, recorder.Body.Bytes())
}
