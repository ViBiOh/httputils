package cache_test

import (
	"context"
	"strconv"
	"testing"

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
	s.Run("success", func() {
		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		got, err := instance.Get(context.Background(), 99679090)
		assert.NoError(s.T(), err)

		expected := getRepository(s.T())

		assert.Equal(s.T(), expected, got)
	})
}
