package snowflake

import (
	"review-service/internal/conf"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func NewSnowFlake(bc *conf.SnowFlake) error {
	st := bc.StartTime
	t, err := time.Parse(time.RFC3339, st)
	if err != nil {
		return err
	}

	snowflake.Epoch = t.UnixMilli()
	snowFakeNode, err := snowflake.NewNode(bc.MachineId)
	if err != nil {
		return err
	}
	node = snowFakeNode
	return nil
}

// 生成雪花算法ID
func GenID() int64 {
	return node.Generate().Int64()
}
