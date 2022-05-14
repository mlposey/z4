package iden

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"time"
)

// TODO: Expand id bytes to fill padding?

type TaskID [16]byte

func New(ts time.Time, index uint64) TaskID {
	var id TaskID
	binary.BigEndian.PutUint64(id[:8], uint64(ts.UTC().UnixNano()))
	binary.BigEndian.PutUint64(id[8:], index)
	return id
}

var (
	// Min is the smallest possible task ID.
	Min TaskID = New(time.Time{}, 0)

	// Max is the largest possible task ID.
	Max TaskID = New(time.Unix(math.MaxInt64, 0), math.MaxUint64)
)

func (t TaskID) Time() time.Time {
	ts := binary.BigEndian.Uint64(t[:8])
	return time.Unix(0, int64(ts))
}

func (t TaskID) Index() uint64 {
	return binary.BigEndian.Uint64(t[8:])
}

func (t TaskID) String() string {
	return base64.StdEncoding.EncodeToString(t[:])
}
