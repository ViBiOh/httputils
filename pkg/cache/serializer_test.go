package cache_test

import (
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func BenchmarkJSONDeserializer(b *testing.B) {
	content := getRepository(b)

	var instance cache.JSONSerializer[Repository]

	payload, err := instance.Encode(content)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = instance.Decode(payload)
	}
}

func BenchmarkStringDeserializer(b *testing.B) {
	var instance cache.StringSerializer

	payload, err := instance.Encode(githubRepoPayload)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = instance.Decode(payload)
	}
}
