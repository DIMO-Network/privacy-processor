package processors

// import (
// 	"encoding/json"
// 	"time"
// )

// type CloudEvent struct {
// 	ID          string    `json:"id"`
// 	Source      string    `json:"source"`
// 	SpecVersion string    `json:"specversion"`
// 	Subject     string    `json:"subject"`
// 	Time        time.Time `json:"time"`
// 	Type        string    `json:"type"`
// }

// type JSONCodec struct {
// 	Factory func() interface{}
// }

// func (c *JSONCodec) Encode(value interface{}) ([]byte, error) {
// 	return json.Marshal(value)
// }

// func (c *JSONCodec) Decode(data []byte) (interface{}, error) {
// 	value := c.Factory()
// 	return value, json.Unmarshal(data, value)
// }
