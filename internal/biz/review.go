package biz

import (
	"context"
	"errors"
	v1 "review-service/api/review/v1"
	"review-service/internal/data/model"
	"review-service/pkg/snowflake"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// 商家回复
type ReviewReply struct {
	ReplyID   int64
	ReviewID  int64
	StoreID   int64
	PicInfo   string
	VideoInfo string
	Content   string
}

// ReviewRepo is a Review repo.
type ReviewRepo interface {
	SaveReview(context.Context, *model.ReviewInfo) (int64, error) // C端
	GetReviewByOrderID(context.Context, int64) (*model.ReviewInfo, error)
	ReplyReview(context.Context, *ReviewReply) (int64, error) // B端
	GetReviewByReviewID(context.Context, int64) (*model.ReviewInfo, error)
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

// 创建评论
func (uc *ReviewUsecase) SaveReview(ctx context.Context, r *model.ReviewInfo) (int64, error) {
	//	1. 业务校验，同一个订单只能创建一次评论
	review, err := uc.repo.GetReviewByOrderID(ctx, r.OrderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		v1.ErrorGormBadErr("订单id:%d查询失败", r.OrderID)
		return 0, v1.ErrorGormBadErr("订单id:%d查询失败", r.OrderID)
	}
	if review != nil {
		return 0, v1.ErrorReviewRepeatedErr("订单id:%d已存在评论，不能重复创建", r.OrderID)
	}

	// 2. reviewID根据雪花算法生成分布式唯一ID
	r.ReviewID = snowflake.GenID()

	// 3. 查看订单信息和商品快照
	// 调用订单相关的rpc接口获取订单信息
	// TODO: 此处省略调用订单服务的代码

	// 4. 评论入库
	return uc.repo.SaveReview(ctx, r)
}

// 回复评论
func (uc *ReviewUsecase) ReplyReview(ctx context.Context, reply *ReviewReply) (int64, error) {
	// 1. 同一条评论只能回复一次
	review, err := uc.repo.GetReviewByReviewID(ctx, reply.ReviewID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.Errorf("评论id:%d查询失败, err:%v", reply.ReviewID, err)
		return 0, v1.ErrorGormBadErr("评论查询失败")
	}

	if review == nil {
		uc.log.Errorf("评论id:%d不存在，无法回复", reply.ReviewID)
		return 0, v1.ErrorGormBadErr("评论不存在，无法回复")
	}

	if review.HasReply == 1 {
		return 0, v1.ErrorReviewHasReplyErr("评论id:%d已回复", reply.ReviewID)
	}

	// 2. 不能水平越权【A商家不能回复B商家下用户的评论】
	if review.StoreID != reply.StoreID {
		uc.log.Errorf("商家id:%d无权限回复评论id:%d", reply.StoreID, reply.ReviewID)
		return 0, v1.ErrorReviewUnauthorizedAccess("水平越权")
	}

	// 3. 回复入库
	reply.ReplyID = snowflake.GenID()
	return uc.repo.ReplyReview(ctx, reply)
}
