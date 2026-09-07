// Package feed builds the JSON document a client receives for one account's
// activity feed.
package feed

import "time"

// Event is one activity record as the service stores it.
type Event struct {
	ID   string
	Kind string
	At   time.Time
}

// Entry is one event as the client sees it. At is RFC 3339.
type Entry struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	At   string `json:"at"`
}

// Document is the payload returned to the client.
//
// The client is a browser bundle with a generated decoder: it walks
// document.events and document.kinds directly. Both members are arrays in
// every response, whatever the account has done, and an account with no
// activity is an ordinary response rather than a special case.
type Document struct {
	Account string   `json:"account"`
	Events  []Entry  `json:"events"`
	Kinds   []string `json:"kinds"`
}

// Build renders the document for account.
//
// Events keeps the order of events. Kinds lists the distinct Kind values
// present, sorted. An event with an empty ID is dropped, and so is its Kind.
func Build(account string, events []Event) Document {
	panic("not implemented")
}
