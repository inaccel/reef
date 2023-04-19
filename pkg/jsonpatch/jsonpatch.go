package jsonpatch

import (
	"encoding/json"

	"github.com/wI2L/jsondiff"
)

func Diff(before, after interface{}) ([]byte, error) {
	beforeRaw, err := json.Marshal(before)
	if err != nil {
		return nil, err
	}
	afterRaw, err := json.Marshal(after)
	if err != nil {
		return nil, err
	}
	operations, err := jsondiff.CompareJSON(beforeRaw, afterRaw)
	if err != nil {
		return nil, err
	}
	return json.Marshal(operations)
}
