package data

import (
	"context"
	"errors"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
)

type appealRepo struct {
	data *Data
	log  *log.Helper
}

func NewAppealRepo(data *Data, logger log.Logger) biz.AppealRepo {
	return &appealRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SaveAppeal 创建申诉
func (r *appealRepo) SaveAppeal(ctx context.Context, appeal *biz.Appeal) (int64, error) {
	err := r.data.query.Transaction(func(tx *query.Query) error {
		err := tx.ReviewAppealInfo.WithContext(ctx).Create(&model.ReviewAppealInfo{
			AppealID:  appeal.AppealID,
			ReviewID:  appeal.ReviewID,
			StoreID:   appeal.StoreID,
			Content:   appeal.Content,
			PicInfo:   appeal.PicInfo,
			VideoInfo: appeal.VideoInfo,
		})
		if err != nil {
			return err
		}
		updateRes, err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(appeal.ReviewID)).Update(tx.ReviewInfo.Status, appeal.Status)
		if err != nil {
			return err
		}
		if updateRes.RowsAffected == 0 {
			return errors.New("更新评论状态失败")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return appeal.AppealID, nil
}

// GetReviewByReviewID 根据reviewID获取评论
func (r *appealRepo) GetReviewByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	review, err := r.data.query.ReviewInfo.WithContext(ctx).Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return nil, err
	}
	return review, nil
}

// UpdateReviewStatus 更新评论状态
func (r *appealRepo) UpdateReviewStatus(ctx context.Context, reviewID int64, status int32) error {
	updateRes, err := r.data.query.ReviewInfo.WithContext(ctx).Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).Update(r.data.query.ReviewInfo.Status, status)
	if err != nil {
		return err
	}
	if updateRes.RowsAffected == 0 {
		return errors.New("更新评论状态失败")
	}
	return nil
}
