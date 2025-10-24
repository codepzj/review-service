package biz

import (
	"context"
	"errors"
	"fmt"
	"review-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// Review is a Review model.
type Review struct {
}

// ReviewRepo is a Review repo.
type ReviewRepo interface {
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByOrderID(context.Context, int64) (*model.ReviewInfo, error)
}

// ReviewUsecase is a Review usecase.
type ReviewUsecase struct {
	repo ReviewRepo
	log  *log.Helper
}

// NewReviewUsecase new a Review usecase.
func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{repo: repo, log: log.NewHelper(logger)}
}

// SaveReview creates a Review, and returns the new Review.
func (uc *ReviewUsecase) SaveReview(ctx context.Context, r *model.ReviewInfo) (*model.ReviewInfo, error) {
	//	1. 业务校验，同一个订单只能创建一次评论
	existingReview, err := uc.repo.GetReviewByOrderID(ctx, r.OrderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("订单id:%d查询失败", r.OrderID)
	}
	if existingReview != nil {
		return nil, fmt.Errorf("订单id:%d已存在评论，不能重复创建", r.OrderID)
	}

	// 2. reviewID根据雪花算法生成分布式唯一ID

	// 3. 查看订单信息和商品快照

	// 4. 评论入库
	return uc.repo.SaveReview(ctx, r)
}
