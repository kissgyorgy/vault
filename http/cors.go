package http

import (
	"net/http"
	"strings"

	"github.com/hashicorp/vault/helper/strutil"
	"github.com/hashicorp/vault/vault"
)

var preflightHeaders = map[string]string{
	"Access-Control-Allow-Headers": "*",
	"Access-Control-Max-Age":       "300",
}

var allowedMethods = []string{
	http.MethodDelete,
	http.MethodGet,
	http.MethodOptions,
	http.MethodPost,
	http.MethodPut,
	"LIST", // LIST is not an official HTTP method, but Vault supports it.
}

func wrapCORSHandler(h http.Handler, core *vault.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		corsConf := core.CORSConfig()

		origin := req.Header.Get("Origin")
		requestMethod := req.Header.Get("Access-Control-Request-Method")

		// If CORS is not enabled or if no Origin header is present (i.e. the request
		// is from the Vault CLI. A browser will always send an Origin header), then
		// just return a 204.
		if !corsConf.IsEnabled() || origin == "" {
			h.ServeHTTP(w, req)
			return
		}

		// Return a 403 if the origin is not
		// allowed to make cross-origin requests.
		if !corsConf.IsValidOrigin(origin) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if req.Method == http.MethodOptions && !strutil.StrListContains(allowedMethods, requestMethod) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")

		// apply headers for preflight requests
		if req.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ","))

			for k, v := range preflightHeaders {
				w.Header().Set(k, v)
			}
			return
		}

		h.ServeHTTP(w, req)
		return
	})
}
