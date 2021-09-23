package jsonpatch

import (
	"encoding/json"

	"github.com/mattbaird/jsonpatch"
)

func Diff(before interface{}, after interface{}) ([]byte, error) {
	beforeRaw, err := json.Marshal(before)
	if err != nil {
		return nil, err
	}
	afterRaw, err := json.Marshal(after)
	if err != nil {
		return nil, err
	}
	operations, err := jsonpatch.CreatePatch(beforeRaw, afterRaw)
	if err != nil {
		return nil, err
	}
	return json.Marshal(operations)
}
