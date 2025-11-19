package generator

import (
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/bwmarrin/snowflake"
)

var idGenerator *IDGenerator

type IDGenerator struct {
	Node *snowflake.Node
}

// 初始化函数 [EN] initialization function
func init() {
	// 创建节点 [EN] Create node
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Errorf("snowflake-generator init err: %s", err.Error())
		panic(err)
	}
	idGenerator = &IDGenerator{Node: node}
}

func GetGenerator() *IDGenerator {
	return idGenerator
}

func (c *IDGenerator) NewID() string {
	return c.Node.Generate().String()
}
