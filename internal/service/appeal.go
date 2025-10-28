package service

import (
	"context"
	"review-service/internal/biz"

	pb "review-service/api/review/v1"
)

type AppealService struct {
	uc *biz.AppealUsecase
}

func NewAppealService(uc *biz.AppealUsecase) *AppealService {
	return &AppealService{uc: uc}
}

// CreateAppeal 创建申诉
func (s *AppealService) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealResponse, error) {
	appealID, err := s.uc.SaveAppeal(ctx, &biz.Appeal{
		ReviewID: req.ReviewId,
		StoreID:  req.StoreId,
		Content:  req.Content,
		PicInfo:  req.PicInfo,
		VideoInfo: req.VideoInfo,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateAppealResponse{AppealId: appealID}, nil
}