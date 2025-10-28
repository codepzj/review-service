package biz

import (
	"strings"
	"time"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewReviewUsecase)

// ReviewInfo 评价表
type ReviewInfo struct {
	ID             int64   `json:"id,string"`
	CreateBy       string  `json:"create_by"`
	UpdateBy       string  `json:"update_by"`
	CreateAt       Mytime  `json:"create_at"`
	UpdateAt       Mytime  `json:"update_at"`
	DeleteAt       *Mytime `json:"delete_at"`
	Version        int32   `json:"version,string"`
	ReviewID       int64   `json:"review_id,string"`
	Content        string  `json:"content"`
	Score          int32   `json:"score,string"`
	ServiceScore   int32   `json:"service_score,string"`
	ExpressScore   int32   `json:"express_score,string"`
	HasMedia       int32   `json:"has_media,string"`
	OrderID        int64   `json:"order_id,string"`
	SkuID          int64   `json:"sku_id,string"`
	SpuID          int64   `json:"spu_id,string"`
	StoreID        int64   `json:"store_id,string"`
	UserID         int64   `json:"user_id,string"`
	Anonymous      int32   `json:"anonymous,string"`
	Tags           string  `json:"tags"`
	PicInfo        string  `json:"pic_info"`
	VideoInfo      string  `json:"video_info"`
	Status         int32   `json:"status,string"`
	IsDefault      int32   `json:"is_default,string"`
	HasReply       int32   `json:"has_reply,string"`
	OpReason       string  `json:"op_reason"`
	OpRemarks      string  `json:"op_remarks"`
	OpUser         string  `json:"op_user"`
	GoodsSnapshoot string  `json:"goods_snapshoot"`
	ExtJSON        string  `json:"ext_json"`
	CtrlJSON       string  `json:"ctrl_json"`
}

type Mytime time.Time

func (mt *Mytime) UnmarshalJSON(data []byte) error {
	// 去除引号
	s := strings.Trim(string(data), "\"")

	// 处理空值、空对象、null等情况
	if s == "" || s == "{}" || s == "null" {
		*mt = Mytime(time.Time{})
		return nil
	}

	// 解析时间
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}
	*mt = Mytime(t)
	return nil
}
