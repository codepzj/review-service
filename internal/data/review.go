package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

// NewReviewRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SaveReview 创建评论
func (r *reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (int64, error) {
	if err := r.data.query.ReviewInfo.WithContext(ctx).Create(review); err != nil {
		return 0, err
	}
	return review.ReviewID, nil
}

// GetReviewByOrderID 根据订单ID获取评论
func (r *reviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) (*model.ReviewInfo, error) {
	reviewInfo := r.data.query.ReviewInfo
	review, err := reviewInfo.WithContext(ctx).Where(reviewInfo.OrderID.Eq(orderID)).First()
	if err != nil {
		return nil, err
	}
	return review, nil
}

// ReplyReview 商家回复评论
func (r *reviewRepo) ReplyReview(ctx context.Context, reply *biz.ReviewReply) (int64, error) {
	reviewReply := &model.ReviewReplyInfo{
		ReplyID:   reply.ReplyID,
		ReviewID:  reply.ReviewID,
		StoreID:   reply.StoreID,
		PicInfo:   reply.PicInfo,
		VideoInfo: reply.VideoInfo,
		Content:   reply.Content,
	}

	// 开启事务
	err := r.data.query.Transaction(func(tx *query.Query) error {
		// 1.商家回复表添加一条记录
		if err := tx.ReviewReplyInfo.WithContext(ctx).Create(reviewReply); err != nil {
			return err
		}

		// 2.评论表更新回复状态
		updateRes, err := tx.ReviewInfo.WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(reply.ReviewID)).
			UpdateColumn(tx.ReviewInfo.HasReply, 1)

		if err != nil {
			return err
		}
		if updateRes.RowsAffected == 0 {
			return errors.New("更新评论已回复失败")
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return reviewReply.ReplyID, nil
}

// GetReviewByReviewID 根据评论ID获取评论
func (r *reviewRepo) GetReviewByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	review := r.data.query.ReviewInfo
	rv, err := review.WithContext(ctx).Where(review.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return nil, err
	}
	return rv, nil
}

// GetReviewListByStoreID 根据店铺ID获取评论列表
func (r *reviewRepo) GetReviewListByStoreID(ctx context.Context, storeID int64, offset int32, size int32) ([]*biz.ReviewInfo, error) {
	resp, err := r.data.esClient.Search().
		Index("review").
		Query(&types.Query{
			Term: map[string]types.TermQuery{
				"store_id": {Value: storeID},
			},
		}).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		From(int(offset)).
		Size(int(size)).
		Do(ctx)
	if err != nil {
		r.log.Errorf("根据店铺ID获取评论列表失败: %v", err)
		return nil, err
	}
	reviews := make([]*biz.ReviewInfo, resp.Hits.Total.Value)
	// 遍历hits，解析评论
	for i, hit := range resp.Hits.Hits {
		var review biz.ReviewInfo
		err = json.Unmarshal(hit.Source_, &review)
		if err != nil {
			r.log.Errorf("解析评论失败: %v", err)
			continue
		}
		reviews[i] = &review
	}
	return reviews, nil
}

var g singleflight.Group

// GetSingleflightReviewListByStoreID singleflight放缓存击穿
func (r *reviewRepo) GetSingleflightReviewListByStoreID(ctx context.Context, storeID int64, offset int32, size int32) ([]*biz.ReviewInfo, error) {
	key := fmt.Sprintf("review:%d:%d:%d", storeID, offset, size)
	val, err, _ := g.Do(key, func() (interface{}, error) {
		// 1. 先从缓存查
		result, err := r.getDataFromRedis(ctx, key)
		if err == nil {
			return result, nil
		}

		// 2. 未命中缓存，直接查es
		if errors.Is(err, redis.Nil) {
			result, err := r.GetReviewListByStoreID(ctx, storeID, offset, size)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(result)
			if err != nil {
				return nil, errors.New("序列化评论列表失败")
			}
			return data, r.setDataToRedis(ctx, key, data)
		}
		// 3. 查不到，说明redis崩了，返回错误
		return nil, nil
	})
	rs := val.([]byte)
	if err != nil {
		return nil, errors.New("获取评论列表失败")
	}
	reviews := []*biz.ReviewInfo{}
	err = json.Unmarshal(rs, &reviews)
	if err != nil {
		return nil, errors.New("解析评论列表失败")
	}
	return reviews, nil
}

func (r *reviewRepo) getDataFromRedis(ctx context.Context, key string) ([]byte, error) {
	return r.data.cache.Get(ctx, key).Bytes()
}

func (r *reviewRepo) setDataToRedis(ctx context.Context, key string, data []byte) error {
	return r.data.cache.Set(ctx, key, data, 1*time.Minute).Err()
}
