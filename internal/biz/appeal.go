package biz

import (
	"context"
	v1 "review-service/api/review/v1"
	"review-service/internal/data/model"
	"review-service/pkg/snowflake"

	"github.com/go-kratos/kratos/v2/log"
)

type Appeal struct {
	AppealID  int64
	ReviewID  int64
	StoreID   int64
	Content   string
	PicInfo   string
	VideoInfo string
	Status    int32 // 审核状态
}

type AppealRepo interface {
	SaveAppeal(context.Context, *Appeal) (int64, error)
	GetReviewByReviewID(context.Context, int64) (*model.ReviewInfo, error)
	UpdateReviewStatus(context.Context, int64, int32) error
}

type AppealUsecase struct {
	repo AppealRepo
	log  *log.Helper
}

func NewAppealUsecase(repo AppealRepo, logger log.Logger) *AppealUsecase {
	return &AppealUsecase{repo: repo, log: log.NewHelper(logger)}
}

// SaveAppeal 创建申诉
func (uc *AppealUsecase) SaveAppeal(ctx context.Context, appeal *Appeal) (int64, error) {
	uc.log.WithContext(ctx).Infof("SaveAppeal: %v", appeal)
	// 1 若评论申诉过，不能重复申诉
	review, err := uc.repo.GetReviewByReviewID(ctx, appeal.ReviewID)
	if err != nil {
		return 0, err
	}
	if review == nil {
		uc.log.WithContext(ctx).Warnf("评论不存在[review_id:%d]", appeal.ReviewID)
		return 0, v1.ErrorGormBadErr("评论不存在")
	}
	if review.Status >= 10 {
		uc.log.WithContext(ctx).Warnf("评论已申诉[review_id:%d]，不能重复申诉", appeal.ReviewID)
		return 0, v1.ErrorReviewAppealedErr("评论已申诉，不能重复申诉")
	}
	// 2 创建申诉记录，并设置评论为待审核状态
	appeal.AppealID = snowflake.GenID()
	appeal.Status = 10
	appealID, err := uc.repo.SaveAppeal(ctx, appeal)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("创建申诉失败[review_id:%d]，%v", appeal.ReviewID, err)
		return 0, err
	}
	return appealID, nil
}
