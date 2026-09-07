// Package feed renders the JSON document a client receives for one account's
// activity feed.
package feed

import "time"

// Event is one activity record as the service stores it.
type Event struct {
	ID   string
	Kind string
	At   time.Time
}

// Render returns the JSON document for account. An empty account is an error
// whose message names the field.
//
// The client is a browser bundle with a generated decoder: it walks the
// members below directly, and every one of them has the same JSON type in
// every response whatever the account has done. An account with no activity is
// an ordinary response rather than a special case.
//
//	{
//	  "account": "acct-2",
//	  "events": [ {"id": "e3", "kind": "logout", "at": "2026-09-07T08:30:00Z"} ],
//	  "kinds": ["login", "logout"],
//	  "counts": {"login": 2, "logout": 1}
//	}
//
// events keeps the order the events were given in, with at formatted as
// RFC 3339. An event with an empty ID never happened: it is absent from
// events, from kinds and from counts. kinds lists the distinct kinds present,
// sorted. counts maps each of those kinds to how many events carried it.
func Render(account string, events []Event) ([]byte, error) {
	panic("not implemented")
}
