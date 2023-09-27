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
	s.Run("no redis", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(context.Background(), 1)
		assert.ErrorContains(s.T(), err, "not implemented")
		assert.Equal(s.T(), Repository{}, got)
	})

	s.Run("bypassed", func() {
		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(cache.Bypass(context.Background()), 1)
		assert.ErrorContains(s.T(), err, "not implemented")
		assert.Equal(s.T(), Repository{}, got)
	})

	s.Run("fetch and store", func() {
		id := 1234567890
		expected := getRepository(s.T())
		expected.ID = id

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

	s.Run("fetch and store with memory", func() {
		id := 987654321
		expected := getRepository(s.T())
		expected.ID = id

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil).WithClientSideCaching(ctx, "fetch_and_store")

		got, err := instance.Get(context.Background(), id)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected, got)

		// Wait for async save
		time.Sleep(time.Second)

		got, err = instance.Get(context.Background(), id)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected, got)
	})

	s.Run("fetch not found", func() {
		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(context.Background(), 1)

		assert.ErrorContains(s.T(), err, "not implemented")
		assert.Equal(s.T(), Repository{}, got)
	})

	s.Run("cache invalid", func() {
		id := 99679090
		expected := getRepository(s.T())

		err := s.integration.Client().Store(context.Background(), strconv.Itoa(id), "{", 0)
		assert.NoError(s.T(), err)

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		got, err := instance.Get(context.Background(), id)

		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected, got)
	})

	s.Run("store error", func() {
		valueFunc := func() string { return "hello" }

		fetchFuncStruct := func(ctx context.Context, id int) (jsonErrorItem, error) {
			return jsonErrorItem{
				ID:    id,
				Value: valueFunc,
			}, nil
		}

		expected, _ := fetchFuncStruct(context.Background(), 1)

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchFuncStruct, nil)

		got, err := instance.Get(context.Background(), 1)

		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expected.ID, got.ID)
		assert.Equal(s.T(), expected.Value(), got.Value())
	})
}
