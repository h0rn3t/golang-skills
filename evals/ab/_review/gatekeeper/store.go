package gatekeeper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"
)

// ErrNotFound reports a document, export, invite, or session that does not exist.
var ErrNotFound = errors.New("not found")

// Principal is the authenticated caller.
type Principal struct {
	UserID string
}

// Doc is one document of an organisation.
type Doc struct {
	ID     string `json:"id"`
	OrgID  string `json:"org_id"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Export is a file rendered from a selection of documents for its owner.
type Export struct {
	ID    string `json:"id"`
	OrgID string `json:"org_id"`
	Owner string `json:"owner"`
	Ready bool   `json:"ready"`
	Data  string `json:"-"`
}

// Invite admits one user to an organisation until it expires.
type Invite struct {
	Code    string
	OrgID   string
	Expires time.Time
	Used    bool
}

// Store keeps documents, exports, invites, and memberships in memory.
type Store struct {
	mu      sync.Mutex
	seq     int
	docs    map[string]Doc
	exports map[string]Export
	invites map[string]Invite
	members map[string]map[string]bool // org ID to member user IDs
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{
		docs:    map[string]Doc{},
		exports: map[string]Export{},
		invites: map[string]Invite{},
		members: map[string]map[string]bool{},
	}
}

func (s *Store) nextID() string {
	s.seq++
	return strconv.Itoa(s.seq)
}

// Doc returns the document with the given ID.
func (s *Store) Doc(id string) (Doc, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.docs[id]
	if !ok {
		return Doc{}, ErrNotFound
	}
	return d, nil
}

// Create stores a new document in org and returns it.
func (s *Store) Create(org, author, title, body string) Doc {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := Doc{ID: s.nextID(), OrgID: org, Author: author, Title: title, Body: body}
	s.docs[d.ID] = d
	return d
}

// AddMember makes user a member of org.
func (s *Store) AddMember(org, user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.members[org] == nil {
		s.members[org] = map[string]bool{}
	}
	s.members[org][user] = true
}

// IsMember reports whether user belongs to org.
func (s *Store) IsMember(org, user string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.members[org][user]
}

// NewExport records a pending export of org for owner.
func (s *Store) NewExport(org, owner string) Export {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := Export{ID: s.nextID(), OrgID: org, Owner: owner}
	s.exports[e.ID] = e
	return e
}

// FinishExport stores the rendered data and marks the export ready.
func (s *Store) FinishExport(id, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.exports[id]
	e.Data, e.Ready = data, true
	s.exports[id] = e
}

// Export returns the export with the given ID.
func (s *Store) Export(id string) (Export, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.exports[id]
	if !ok {
		return Export{}, ErrNotFound
	}
	return e, nil
}

// AddInvite stores inv under its code.
func (s *Store) AddInvite(inv Invite) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invites[inv.Code] = inv
}

// Invite returns the invite with the given code.
func (s *Store) Invite(code string) (Invite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[code]
	if !ok {
		return Invite{}, ErrNotFound
	}
	return inv, nil
}

// MarkUsed spends the invite with the given code.
func (s *Store) MarkUsed(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv := s.invites[code]
	inv.Used = true
	s.invites[code] = inv
}

// Sessions maps the SHA-256 of a session ID to its principal, so a copy of
// the table holds no usable cookie. Revoked asks the shared revocation list,
// a network call that can fail.
type Sessions struct {
	mu      sync.Mutex
	byHash  map[string]Principal
	Revoked func(ctx context.Context, hash string) (bool, error)
}

// NewSessions returns an empty table that checks revocation with revoked.
func NewSessions(revoked func(ctx context.Context, hash string) (bool, error)) *Sessions {
	return &Sessions{byHash: map[string]Principal{}, Revoked: revoked}
}

// Start opens a session for user and returns the cookie value.
func (s *Sessions) Start(user string) string {
	id := rand.Text()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byHash[hashID(id)] = Principal{UserID: user}
	return id
}

// Lookup returns the principal of the session whose ID hashes to hash.
func (s *Sessions) Lookup(hash string) (Principal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.byHash[hash]
	if !ok {
		return Principal{}, ErrNotFound
	}
	return p, nil
}

func hashID(id string) string {
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}
