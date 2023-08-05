package cache

import "encoding/json"

type Serializer[V any] interface {
	Encode(V) ([]byte, error)
	Decode([]byte) (V, error)
}

type JSONSerializer[V any] struct{}

func (JSONSerializer[V]) Encode(payload V) ([]byte, error) {
	return json.Marshal(payload)
}

func (JSONSerializer[V]) Decode(payload []byte) (V, error) {
	var content V
	return content, json.Unmarshal(payload, &content)
}

type StringSerializer struct{}

func (StringSerializer) Encode(payload string) ([]byte, error) {
	return []byte(payload), nil
}

func (StringSerializer) Decode(payload []byte) (string, error) {
	return string(payload), nil
}
