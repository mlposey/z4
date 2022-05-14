package iden

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

var ErrNoTime = errors.New("id contains no time component")

// TODO: Expand id bytes to fill padding?

type TaskID [16]byte

func New(ts time.Time, index uint64) TaskID {
	var id TaskID
	if !ts.IsZero() {
		binary.BigEndian.PutUint64(id[:8], uint64(ts.UTC().UnixNano()))
	}
	binary.BigEndian.PutUint64(id[8:], index)
	return id
}

var (
	// Min is the smallest possible task ID.
	Min TaskID = New(time.Time{}, 0)

	// Max is the largest possible task ID.
	Max TaskID = New(time.Unix(math.MaxInt64, 0), math.MaxUint64)
)

func (t TaskID) MustTime() time.Time {
	ts := binary.BigEndian.Uint64(t[:8])
	return time.Unix(0, int64(ts))
}

func (t TaskID) Time() (time.Time, error) {
	ts := binary.BigEndian.Uint64(t[:8])
	if ts == 0 {
		return time.Time{}, ErrNoTime
	}
	return time.Unix(0, int64(ts)), nil
}

func (t TaskID) Index() uint64 {
	return binary.BigEndian.Uint64(t[8:])
}

func (t TaskID) String() string {
	return base64.StdEncoding.EncodeToString(t[:])
}
