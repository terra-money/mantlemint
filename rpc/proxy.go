package rpc

import (
	"github.com/tendermint/tendermint/libs/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
)

type ProxyMiddleware struct {
	lcdUrls []string
	proxies []*httputil.ReverseProxy
}

func NewProxyMiddleware(lcdUrls []string) ProxyMiddleware {
	var proxies []*httputil.ReverseProxy
	for _, u := range lcdUrls {
		lcdUrl, err := url.Parse(u)
		if err != nil {
			panic(err)
		}
		proxies = append(proxies, httputil.NewSingleHostReverseProxy(lcdUrl))
	}
	return ProxyMiddleware{
		lcdUrls: lcdUrls,
		proxies: proxies,
	}
}

func (pm ProxyMiddleware) HandleRequest(writer http.ResponseWriter, request *http.Request, handler http.Handler) {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		// randomly pick a proxy from the list
		proxyToUse := rand.NewRand().Intn(len(pm.proxies))
		pm.proxies[proxyToUse].ServeHTTP(writer, request)
		return
	}
	writer.WriteHeader(recorder.Code)
	writer.Write(recorder.Body.Bytes())
	return
}
