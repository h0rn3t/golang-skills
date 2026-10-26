// Package zoning assigns a tariff zone to a metering point over time, from the
// point's recorded property history.
package zoning

import (
	"context"
	"time"
)

// Period is the half-open interval [From, To).
type Period struct {
	From, To time.Time
}

// Snapshot is one stretch of a point's recorded properties, valid over Period.
//
// Props holds the values exactly as the history service decodes them from
// JSON: a string, a float64, a bool, or nil. A key that is absent and a key
// whose value is nil both mean the property was not recorded.
type Snapshot struct {
	Period Period
	Props  map[string]any
}

// Store is the point's property history and its zone column.
type Store interface {
	// History returns the snapshots of point that overlap period, sorted by
	// start and never overlapping each other. A snapshot may extend past
	// period, and the snapshots may leave gaps.
	History(ctx context.Context, point string, period Period) ([]Snapshot, error)
	// SetZone records zone for point over period; nil clears it.
	SetZone(ctx context.Context, point string, period Period, zone *string) error
}

// Request is one bind or unbind of a zone.
type Request struct {
	Point  string
	Period Period
	Zone   string
	// Bind is true to bind Zone and false to unbind it.
	Bind bool
	// Static skips the property check.
	Static bool
	// Match lists, by key, the property values a snapshot must carry for the
	// zone to apply over it. An empty Match matches every snapshot.
	Match map[string]string
}

// Apply applies req to the point's zone column and reports whether it
// processed anything.
//
// Processing starts no earlier than launch, the day zoning went live: the
// processed period is req.Period with its start raised to launch. When that
// leaves nothing, because the processed period would end at or before it
// starts, Apply returns false and calls neither History nor SetZone.
//
// A static request sets the zone over the whole processed period in one call,
// Zone on bind and nil on unbind, without reading History.
//
// Any other request reads History over the processed period once and sets the
// zone over each piece of that period:
//
//   - a piece covered by a matching snapshot gets Zone on bind and nil on unbind;
//   - a piece covered by a snapshot that does not match gets nil;
//   - a piece no snapshot covers gets nil on bind and is left alone on unbind.
//
// A snapshot matches when every key in Match is recorded on it with the same
// value. Values compare as text: a string as is, a float64 in its shortest
// decimal form (220, 0.4), a bool as true or false. The "kind" property was
// never recorded on older points, so on a snapshot that does not record it,
// kind reads as "site".
//
// SetZone is called once per piece, in ascending order of piece start, with
// every piece clipped to the processed period. Adjacent pieces are never
// merged.
//
// A recorded value of any other type, such as a list or an object, fails the
// whole request with an error naming the property and the point, before any
// SetZone call. A History or SetZone failure is returned wrapped; Apply stops
// at the first SetZone failure. On success Apply returns true.
func Apply(ctx context.Context, st Store, launch time.Time, req Request) (bool, error) {
	panic("not implemented")
}
