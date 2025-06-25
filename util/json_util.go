package util

import "encoding/json"

func JsonInterfaceToString(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(b)
}
