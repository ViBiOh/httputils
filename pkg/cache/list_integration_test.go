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

type ListSuite struct {
	suite.Suite

	integration *test.RedisIntegration
}

func (ls *ListSuite) SetupSuite() {
	ls.integration = test.NewRedisIntegration(ls.T())
	ls.integration.Bootstrap("cache_list")
}

func (ls *ListSuite) TearDownSuite() {
	ls.integration.Close()
}

func (ls *ListSuite) TearDownTest() {
	ls.integration.Reset()
}

func TestListSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	suite.Run(t, new(ListSuite))
}

func (s *ListSuite) TestList() {
	s.Run("no item", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, noFetch, nil)

		got, err := instance.List(context.Background())
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository(nil), got)
	})

	s.Run("bypassed", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		got, err := instance.List(cache.Bypass(context.Background()), 10, 20)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second}, got)
	})

	s.Run("no memory no redis", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		got, err := instance.List(context.Background(), 10, 20)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second}, got)
	})

	s.Run("memory no redis", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil).
			WithClientSideCaching(ctx, "memory_no_redis")

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		third := getRepository(s.T())
		third.ID = 30

		got, err := instance.List(context.Background(), 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		got, err = instance.List(context.Background(), 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)
	})

	s.Run("no memory and redis", func() {
		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		third := getRepository(s.T())
		third.ID = 30

		got, err := instance.List(context.Background(), 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		got, err = instance.List(context.Background(), 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)
	})

	s.Run("memory and redis", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil).
			WithClientSideCaching(ctx, "memory_redis")

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		third := getRepository(s.T())
		third.ID = 30

		four := getRepository(s.T())
		four.ID = 40

		got, err := instance.List(context.Background(), 10)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first}, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		err = s.integration.Client().Store(ctx, "20", "invalid_payload", 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Millisecond * 50)

		payload, err := cache.JSONSerializer[Repository]{}.Encode(four)
		assert.NoError(s.T(), err)

		err = s.integration.Client().Store(ctx, "40", payload, 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), 10, 40, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, four, second, third}, got)
	})
}
