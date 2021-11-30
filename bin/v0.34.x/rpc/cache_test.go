package rpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheBackend(t *testing.T) {
	cb := NewCacheBackend(1, "test")

	cb.Set("key", 200, []byte("hello world"))
	cached := cb.Get("key")
	assert.Equal(t, 200, cached.status)
	assert.Equal(t, []byte("hello world"), cached.body)

	cb.Set("key2", 501, []byte("error"))
	cached2 := cb.Get("key2")
	assert.Equal(t, 501, cached2.status)
	assert.Equal(t, []byte("error"), cached2.body)

	testReq := httptest.NewRequest(
		"get",
		"/test/request?param=1",
		nil,
	)
	testRes := httptest.NewRecorder()
	var callCount = 0

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		callCount++
		writer.WriteHeader(123)
		writer.Write([]byte("asdf"))
	})

	// call 3 times
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)

	fmt.Println(callCount)
	assert.Equal(t, 1, callCount)

	cb.Purge()

	callCount = 0
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)
	cb.HandleCachedHTTP(testRes, testReq, handler)

	fmt.Println(callCount)
	assert.Equal(t, callCount, 1)

}
