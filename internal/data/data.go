package data

import (
	"context"
	"review-service/internal/conf"
	"review-service/internal/data/query"

	es "github.com/elastic/go-elasticsearch/v9"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB, NewRedis, NewEsClient)

// Data .
type Data struct {
	query    *query.Query
	cache    *redis.Client
	esClient *es.TypedClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, cache *redis.Client, esClient *es.TypedClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db) // 指定数据库
	return &Data{query: query.Q, cache: cache, esClient: esClient}, cleanup, nil
}

func NewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		log.Fatalf("database start failed: %#v", err)
		panic(err)
	}
	return db
}

func NewRedis(c *conf.Data) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
	})
	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("redis start failed: %#v", err)
		panic(err)
	}
	return client
}

func NewEsClient(config *conf.Elasticsearch) *es.TypedClient {
	esClient, err := es.NewTypedClient(es.Config{
		Addresses: config.Addresses,
	})
	if err != nil {
		log.Fatalf("elasticsearch start failed: %#v", err)
		panic(err)
	}
	return esClient
}
