package utils

import (
	"encoding/json"
)

func InterfaceToStruct(m interface{}, s interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, s)
	if err != nil {
		return err
	}
	return nil
}
