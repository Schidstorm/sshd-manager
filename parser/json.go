package parser

import "encoding/json"

func SerializeJson(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
