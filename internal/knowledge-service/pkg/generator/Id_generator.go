package generator

import (
	"errors"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/bwmarrin/snowflake"
)

var idGenerator = IDGenerator{}

type IDGenerator struct {
	Node *snowflake.Node
}

func init() {
	pkg.AddContainer(idGenerator)
}

func (c IDGenerator) LoadType() string {
	return "snowflake-generator"
}

func (c IDGenerator) Load() error {
	node, err := snowflake.NewNode(1) // 创建节点 [EN] Create node
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("snowflake-generator node is nil")
	}
	idGenerator.Node = node
	return nil
}

func (c IDGenerator) Stop() error {
	return nil
}

func (c IDGenerator) StopPriority() int {
	return pkg.DefaultPriority
}

func (c IDGenerator) NewID() string {
	return idGenerator.Node.Generate().String()
}

func GetGenerator() IDGenerator {
	return idGenerator
}
