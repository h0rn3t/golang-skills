// Package gateway serves the public read-only accounts API.
package gateway

import "net/http"

// Account is one record the API serves.
type Account struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Cents  int64  `json:"cents"`
	Active bool   `json:"active"`
}

// NewServer returns the server for the public accounts API, listening on addr
// and serving the given accounts. Starting and stopping it is the caller's job.
//
// This server is the edge: it is published straight to the open internet with
// no proxy of any kind in front of it. Every client is anonymous, some are on
// mobile links that stall mid-request, many hold connections open long after
// they stop asking for anything, and a share of the traffic is there to find
// out what the process can be made to hold onto.
//
// It answers these requests and nothing else:
//
//	GET /healthz         200, body "ok"
//	GET /accounts        200, a JSON array of every account, ordered by ID
//	GET /accounts/{id}   200, the JSON object for that account; 404 if unknown
//
// GET /accounts also takes one optional query parameter, active, which is
// either "true" or "false" and restricts the array to accounts in that state.
// Any other value for it is a 400 whose body names the parameter.
//
// Any other path is a 404. A known path asked for with any method other than
// GET is a 405.
func NewServer(addr string, accounts []Account) *http.Server {
	panic("not implemented")
}
