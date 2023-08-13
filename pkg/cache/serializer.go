package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

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

type GobSerializer[V any] struct{}

func (GobSerializer[V]) Encode(payload V) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buffer).Encode(payload)

	return buffer.Bytes(), err
}

func (GobSerializer[V]) Decode(payload []byte) (V, error) {
	var content V
	return content, gob.NewDecoder(bytes.NewReader(payload)).Decode(&content)
}

type StringSerializer struct{}

func (StringSerializer) Encode(payload string) ([]byte, error) {
	return []byte(payload), nil
}

func (StringSerializer) Decode(payload []byte) (string, error) {
	return string(payload), nil
}
