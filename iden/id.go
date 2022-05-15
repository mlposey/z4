package iden

import (
	"encoding/binary"
	"errors"
	"math"
	"math/big"
	"time"
)

var ErrNoTime = errors.New("id contains no time component")

type TaskID [16]byte

func New(ts time.Time, index uint64) TaskID {
	var id TaskID
	binary.BigEndian.PutUint64(id[:8], uint64(ts.UnixNano()))
	binary.BigEndian.PutUint64(id[8:], index)
	return id
}

var (
	// Min is the smallest possible task ID.
	Min TaskID = New(time.Unix(0, 0), 0)

	// Max is the largest possible task ID.
	Max TaskID = New(time.Unix(math.MaxInt64, 0).UTC(), math.MaxUint64)

	zeroTime = time.Unix(0, time.Time{}.UnixNano())
)

func (t TaskID) MustTime() time.Time {
	ts := binary.BigEndian.Uint64(t[:8])
	return time.Unix(0, int64(ts))
}

func (t TaskID) Time() (time.Time, error) {
	ts := binary.BigEndian.Uint64(t[:8])
	if time.Unix(0, int64(ts)).Equal(zeroTime) {
		return time.Time{}, ErrNoTime
	}
	return time.Unix(0, int64(ts)), nil
}

func (t TaskID) Index() uint64 {
	return binary.BigEndian.Uint64(t[8:])
}

func (t TaskID) String() string {
	var i big.Int
	i.SetBytes(t[:])
	return i.Text(62)
}
