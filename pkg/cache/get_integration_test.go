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

type GetSuite struct {
	suite.Suite

	integration *test.RedisIntegration
}

func (gs *GetSuite) SetupSuite() {
	gs.integration = test.NewRedisIntegration(gs.T())
	gs.integration.Bootstrap("cache_get")
}

func (gs *GetSuite) TearDownSuite() {
	gs.integration.Close()
}

func (gs *GetSuite) TearDownTest() {
	gs.integration.Reset()
}

func TestGetSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	suite.Run(t, new(GetSuite))
}

func (gs *GetSuite) TestGet() {
	gs.Run("no redis", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(context.Background(), 1)
		assert.ErrorContains(gs.T(), err, "not implemented")
		assert.Equal(gs.T(), Repository{}, got)
	})

	gs.Run("bypassed", func() {
		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(cache.Bypass(context.Background()), 1)
		assert.ErrorContains(gs.T(), err, "not implemented")
		assert.Equal(gs.T(), Repository{}, got)
	})

	gs.Run("fetch and store", func() {
		id := 1234567890
		expected := getRepository(gs.T())
		expected.ID = id

		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		got, err := instance.Get(context.Background(), id)
		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		instance = cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err = instance.Get(context.Background(), id)
		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected, got)
	})

	gs.Run("fetch and store with memory", func() {
		id := 987654321
		expected := getRepository(gs.T())
		expected.ID = id

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil).WithClientSideCaching(ctx, "fetch_and_store")

		got, err := instance.Get(context.Background(), id)
		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		got, err = instance.Get(context.Background(), id)
		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected, got)
	})

	gs.Run("fetch not found", func() {
		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.Get(context.Background(), 1)

		assert.ErrorContains(gs.T(), err, "not implemented")
		assert.Equal(gs.T(), Repository{}, got)
	})

	gs.Run("cache invalid", func() {
		id := 99679090
		expected := getRepository(gs.T())

		err := gs.integration.Client().Store(context.Background(), strconv.Itoa(id), "{", 0)
		assert.NoError(gs.T(), err)

		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		got, err := instance.Get(context.Background(), id)

		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected, got)
	})

	gs.Run("store error", func() {
		valueFunc := func() string { return "hello" }

		fetchFuncStruct := func(ctx context.Context, id int) (jsonErrorItem, error) {
			return jsonErrorItem{
				ID:    id,
				Value: valueFunc,
			}, nil
		}

		expected, _ := fetchFuncStruct(context.Background(), 1)

		instance := cache.New(gs.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchFuncStruct, nil)

		got, err := instance.Get(context.Background(), 1)

		assert.NoError(gs.T(), err)
		assert.Equal(gs.T(), expected.ID, got.ID)
		assert.Equal(gs.T(), expected.Value(), got.Value())
	})
}
