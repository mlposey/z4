package iden

import (
	"fmt"
	"math/big"
)

func ParseString(v string) (TaskID, error) {
	var i big.Int
	_, ok := i.SetString(v, 62)
	if !ok {
		return TaskID{}, fmt.Errorf("invalid task id: %s", v)
	}

	var id TaskID
	copy(id[:], i.Bytes())
	return id, nil
}

func MustParseString(v string) TaskID {
	id, err := ParseString(v)
	if err != nil {
		panic(err)
	}
	return id
}
