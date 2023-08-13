package cache_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	integration *test.RedisIntegration
}

func (r *Suite) SetupSuite() {
	r.integration = test.NewRedisIntegration(r.T())
	r.integration.Bootstrap("cache")
}

func (r *Suite) TearDownSuite() {
	r.integration.Close()
}

func (r *Suite) TearDownTest() {
	r.integration.Reset()
}

func TestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	suite.Run(t, new(Suite))
}

func (s *Suite) TestGet() {
	s.Run("fetch and store", func() {
		id := 99679090
		expected := getRepository(s.T())

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		got, err := instance.Get(context.Background(), id)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected, got)

		// Wait for async save
		time.Sleep(time.Second)

		instance = cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err = instance.Get(context.Background(), id)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected, got)
	})
}
