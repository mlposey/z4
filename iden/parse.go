package iden

import (
	"encoding/base64"
	"fmt"
)

func ParseString(v string) (TaskID, error) {
	dec, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return TaskID{}, err
	}

	if len(dec) != 16 {
		return TaskID{}, fmt.Errorf("invalid task id: %s", v)
	}

	var id TaskID
	copy(id[:], dec)
	return id, nil
}

func MustParseString(v string) TaskID {
	id, err := ParseString(v)
	if err != nil {
		panic(err)
	}
	return id
}
