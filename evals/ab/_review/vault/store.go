// Package vault stores the files a tenant uploads and serves them back: each
// tenant has its own directory under the root, every file carries a stable
// ETag so browsers cache it, and a per-tenant quota bounds what is kept.
package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// ErrBadName reports a file name that is empty or would resolve outside the
// tenant's directory.
var ErrBadName = errors.New("vault: bad file name")

// ErrExists reports a name the tenant has already stored.
var ErrExists = errors.New("vault: file exists")

// ErrQuota reports an upload that would take the tenant over its quota.
var ErrQuota = errors.New("vault: quota exceeded")

// Meta describes one stored file.
type Meta struct {
	Name        string            `json:"name"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	ModTime     time.Time         `json:"mod_time"`
	Expires     time.Time         `json:"expires,omitzero"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// ETag returns the entity tag of the file m describes: a hash of its size,
// modification time and labels that comes out the same on every node, so a
// browser holding the file is told so with a 304.
func (m Meta) ETag() string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, "%d\n%d\n", m.Size, m.ModTime.UnixNano()))
	for k, v := range m.Labels {
		h.Write([]byte(k + "=" + v + "\n"))
	}
	return `"` + hex.EncodeToString(h.Sum(nil))[:32] + `"`
}

// Store keeps each tenant's files under root/<tenant>/ and, in memory, what
// it has learned about them since it started; older files say for themselves.
type Store struct {
	root  string
	quota int64

	mu    sync.Mutex
	used  map[string]int64           // bytes stored, by tenant
	files map[string]map[string]Meta // by tenant, then name
}

// NewStore returns a store over root that allows quota bytes per tenant.
func NewStore(root string, quota int64) *Store {
	return &Store{root: root, quota: quota, used: make(map[string]int64), files: make(map[string]map[string]Meta)}
}

// path returns where name lives inside tenant's directory and refuses names
// that would resolve outside it.
func (s *Store) path(tenant, name string) (string, error) {
	if tenant == "" || name == "" {
		return "", ErrBadName
	}
	dir := filepath.Join(s.root, tenant)
	p := filepath.Clean(filepath.Join(dir, name))
	if !strings.HasPrefix(p, dir) {
		return "", ErrBadName
	}
	return p, nil
}

// remaining returns how many more bytes tenant may store.
func (s *Store) remaining(tenant string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.quota - s.used[tenant]
}

// Put stores the contents of r as tenant's file name, labeled with labels,
// and returns what it stored. A name may be stored once; ErrQuota is
// returned, and nothing kept, when the file would take the tenant over its
// quota.
func (s *Store) Put(tenant, name, contentType string, labels map[string]string, r io.Reader) (Meta, error) {
	p, err := s.path(tenant, name)
	if err != nil {
		return Meta{}, err
	}
	remaining := s.remaining(tenant)
	if remaining <= 0 {
		return Meta{}, ErrQuota
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return Meta{}, fmt.Errorf("put %s/%s: %w", tenant, name, err)
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return Meta{}, ErrExists
	}
	if err != nil {
		return Meta{}, fmt.Errorf("put %s/%s: %w", tenant, name, err)
	}
	n, err := io.Copy(f, io.LimitReader(r, remaining+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > remaining {
		err = ErrQuota
	}
	if err != nil {
		_ = os.Remove(p)
		return Meta{}, fmt.Errorf("put %s/%s: %w", tenant, name, err)
	}
	m := Meta{Name: name, Size: n, ContentType: contentType, ModTime: time.Now(), Labels: labels}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.used[tenant] += n
	if s.files[tenant] == nil {
		s.files[tenant] = make(map[string]Meta)
	}
	s.files[tenant][name] = m
	return m, nil
}

// Open returns tenant's file name for reading, with what is known about it.
func (s *Store) Open(tenant, name string) (*os.File, Meta, error) {
	p, err := s.path(tenant, name)
	if err != nil {
		return nil, Meta{}, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, Meta{}, fmt.Errorf("open %s/%s: %w", tenant, name, err)
	}
	s.mu.Lock()
	m, ok := s.files[tenant][name]
	s.mu.Unlock()
	if !ok {
		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, Meta{}, fmt.Errorf("open %s/%s: %w", tenant, name, err)
		}
		m = Meta{Name: name, Size: fi.Size(), ContentType: "application/octet-stream", ModTime: fi.ModTime()}
	}
	return f, m, nil
}

// Delete removes tenant's file name. Deleting a name that is not stored is
// not an error.
func (s *Store) Delete(tenant, name string) error {
	p, err := s.path(tenant, name)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete %s/%s: %w", tenant, name, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files[tenant], name)
	return nil
}

// List returns what tenant has stored, by name.
func (s *Store) List(tenant string) []Meta {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := slices.SortedFunc(maps.Values(s.files[tenant]), func(a, b Meta) int { return strings.Compare(a.Name, b.Name) })
	if out == nil {
		out = []Meta{}
	}
	return out
}

// Used returns how many bytes tenant has stored.
func (s *Store) Used(tenant string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.used[tenant]
}
