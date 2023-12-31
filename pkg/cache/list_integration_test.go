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

		got, err := instance.List(context.Background(), nil)
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository(nil), got)
	})

	s.Run("bypassed", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		got, err := instance.List(cache.Bypass(context.Background()), nil, 10, 20)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second}, got)
	})

	s.Run("no memory no redis", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		got, err := instance.List(context.Background(), nil, 10, 20)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second}, got)
	})

	s.Run("no memory no redis, load many", func() {
		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchRepository, nil).
			WithMissMany(fetchRepositories)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		got, err := instance.List(context.Background(), nil, 10, 20)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second}, got)
	})

	s.Run("memory no redis", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(nil, func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil).
			WithClientSideCaching(ctx, "memory_no_redis", 10)

		first := getRepository(s.T())
		first.ID = 10

		second := getRepository(s.T())
		second.ID = 20

		third := getRepository(s.T())
		third.ID = 30

		got, err := instance.List(context.Background(), nil, 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Second)

		got, err = instance.List(context.Background(), nil, 10, 20, 30)

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

		got, err := instance.List(context.Background(), nil, 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Second)

		got, err = instance.List(context.Background(), nil, 10, 20, 30)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)
	})

	s.Run("memory and redis", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil).
			WithClientSideCaching(ctx, "memory_redis", 10)

		first := getRepository(s.T())
		first.ID = 1000

		second := getRepository(s.T())
		second.ID = 2000

		third := getRepository(s.T())
		third.ID = 3000

		four := getRepository(s.T())
		four.ID = 4000

		got, err := instance.List(context.Background(), nil, 1000)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first}, got)

		// Wait for async save
		time.Sleep(time.Second)

		err = s.integration.Client().Store(ctx, "2000", "invalid_payload", 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), nil, 1000, 2000, 3000)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Second)

		payload, err := cache.JSONSerializer[Repository]{}.Encode(four)
		assert.NoError(s.T(), err)

		err = s.integration.Client().Store(ctx, "4000", payload, 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), nil, 1000, 4000, 2000, 3000)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, four, second, third}, got)
	})

	s.Run("memory, redis, many and extend", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := cache.New(s.integration.Client(), func(id int) string { return strconv.Itoa(id) }, fetchOnce(), nil).
			WithClientSideCaching(ctx, "memory_redis_may_extend", 10).
			WithTTL(time.Hour).
			WithExtendOnHit(ctx, time.Hour/4, 10).
			WithMissMany(fetchRepositoriesOnce())

		first := getRepository(s.T())
		first.ID = 100

		second := getRepository(s.T())
		second.ID = 200

		third := getRepository(s.T())
		third.ID = 300

		four := getRepository(s.T())
		four.ID = 400

		got, err := instance.List(context.Background(), nil, 100)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first}, got)

		// Wait for async save
		time.Sleep(time.Second)

		err = s.integration.Client().Store(ctx, "200", "invalid_payload", 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), nil, 100, 200, 300)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, second, third}, got)

		// Wait for async save
		time.Sleep(time.Second)

		payload, err := cache.JSONSerializer[Repository]{}.Encode(four)
		assert.NoError(s.T(), err)

		err = s.integration.Client().Store(ctx, "400", payload, 0)
		assert.NoError(s.T(), err)

		got, err = instance.List(context.Background(), nil, 100, 400, 200, 300)

		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []Repository{first, four, second, third}, got)
	})
}
