package service

import (
	"context"

	pb "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"
)

type ReviewService struct {
	pb.UnimplementedReviewServer
	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{
		uc: uc,
	}
}

// 创建回复
func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	reviewID, err := s.uc.SaveReview(ctx, &model.ReviewInfo{
		UserID:       req.UserId,
		OrderID:      req.OrderId,
		StoreID:      req.StoreId,
		PicInfo:      req.PicInfo,
		VideoInfo:    req.VideoInfo,
		Content:      req.Content,
		Score:        req.Score,
		ServiceScore: req.ServiceScore,
		ExpressScore: req.ExpressScore,
		Anonymous:    req.Anonymous,
		Status:       0,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateReviewReply{ReviewId: reviewID}, nil
}

// 商家评论回复
func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReviewReplyRequest) (*pb.ReviewReplyResponse, error) {
	replyID, err := s.uc.ReplyReview(ctx, &biz.ReviewReply{
		ReviewID:  req.ReviewId,
		StoreID:   req.StoreId,
		PicInfo:   req.PicInfo,
		VideoInfo: req.VideoInfo,
		Content:   req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ReviewReplyResponse{ReplyId: replyID}, nil
}
