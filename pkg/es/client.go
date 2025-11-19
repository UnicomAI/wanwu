package es

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	Address  string `json:"address" mapstructure:"address"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

type client struct {
	ctx context.Context
	cli *elasticsearch.Client

	mutex   sync.Mutex
	stopped bool
	stop    chan struct{}
}

func newClient(ctx context.Context, c Config) (*client, error) {
	// 智能判断协议，如果地址没有协议前缀，则尝试HTTPS，失败后尝试HTTP [EN] Intelligently determine the protocol. If the address does not have a protocol prefix, HTTPS will be tried. If it fails, HTTP will be tried.
	addresses := []string{}

	// 如果地址已经包含协议，直接使用 [EN] If the address already contains a protocol, use it directly
	if strings.HasPrefix(c.Address, "http://") || strings.HasPrefix(c.Address, "https://") {
		addresses = append(addresses, c.Address)
	} else {
		// 优先尝试HTTPS，然后HTTP [EN] Try HTTPS first, then HTTP
		addresses = append(addresses, "https://"+c.Address, "http://"+c.Address)
	}

	var lastErr error

	// 尝试每个地址 [EN] try every address
	for _, addr := range addresses {
		cfg := elasticsearch.Config{
			Addresses: []string{addr},
			Username:  c.Username,
			Password:  c.Password,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		esClient, err := elasticsearch.NewClient(cfg)
		if err != nil {
			lastErr = fmt.Errorf("创建ES客户端失败 [%s]: %v", addr, err)
			log.Warnf("创建ES客户端失败，地址: %s, 错误: %v", addr, err)
			continue
		}

		// 测试连接 [EN] test connection
		res, err := esClient.Info()
		if err != nil {
			lastErr = fmt.Errorf("ES连接测试失败 [%s]: %v", addr, err)
			log.Warnf("ES连接测试失败，地址: %s, 错误: %v", addr, err)
			continue
		}

		if res != nil {
			defer res.Body.Close()

			if res.IsError() {
				lastErr = fmt.Errorf("ES连接响应错误 [%s]: %s", addr, res.String())
				log.Warnf("ES连接响应错误，地址: %s, 响应: %s", addr, res.String())
				continue
			}
		}

		log.Infof("ES连接成功，地址: %s", addr)
		return &client{
			ctx:  ctx,
			cli:  esClient,
			stop: make(chan struct{}, 1),
		}, nil
	}

	// 所有地址都失败了 [EN] All addresses failed
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("无法连接到ES，尝试的地址: %v", addresses)
}

func (c *client) Stop() {
	c.mutex.Lock()
	if c.stopped {
		log.Errorf("ES客户端已经停止")
		c.mutex.Unlock()
		return
	}
	c.stopped = true
	close(c.stop)
	c.mutex.Unlock()
	log.Infof("ES客户端停止")
}

func (c *client) Cli() *elasticsearch.Client {
	return c.cli
}

// 写入数据到指定索引 [EN] Write data to the specified index
func (c *client) IndexDocument(ctx context.Context, index string, document interface{}) error {
	docJSON, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %v", err)
	}

	res, err := c.cli.Index(
		index,
		strings.NewReader(string(docJSON)),
		c.cli.Index.WithContext(ctx),
		c.cli.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("写入ES失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES写入响应错误: %s", res.String())
	}

	log.Infof("成功写入ES，索引: %s", index)
	return nil
}

// 根据指定字段条件查询所有数据 [EN] Query all data based on specified field conditions
func (c *client) SearchByFields(ctx context.Context, index string, fieldConditions map[string]interface{}, from, size int) ([]json.RawMessage, int64, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": buildMustQuery(fieldConditions),
			},
		},
		"from": from,
		"size": size,
		"sort": []map[string]interface{}{
			{
				"createdAt": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化查询失败: %v", err)
	}

	res, err := c.cli.Search(
		c.cli.Search.WithContext(ctx),
		c.cli.Search.WithIndex(index),
		c.cli.Search.WithBody(strings.NewReader(string(queryJSON))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("ES查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("ES查询响应错误: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("解析查询结果失败: %v", err)
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("无效的查询结果格式")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("无效的总数格式")
	}

	totalValue, ok := total["value"].(float64)
	if !ok {
		return nil, 0, fmt.Errorf("无效的总数值")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("无效的命中列表格式")
	}

	var documents []json.RawMessage
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		source, ok := hitMap["_source"]
		if !ok {
			continue
		}
		sourceJSON, err := json.Marshal(source)
		if err != nil {
			continue
		}
		documents = append(documents, sourceJSON)
	}

	log.Infof("ES查询成功，索引: %s, 总数: %d, 返回: %d", index, int64(totalValue), len(documents))
	return documents, int64(totalValue), nil
}

// 创建索引模板 [EN] Create index template
func (c *client) CreateIndexTemplate(ctx context.Context, templateName string, templateBody string) error {
	res, err := c.cli.Indices.PutIndexTemplate(
		templateName,
		strings.NewReader(templateBody),
		c.cli.Indices.PutIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("创建索引模板失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引模板响应错误: %s", res.String())
	}

	log.Infof("成功创建索引模板: %s", templateName)
	return nil
}

// 检查索引模板是否存在 [EN] Check if index template exists
func (c *client) IndexTemplateExists(ctx context.Context, templateName string) (bool, error) {
	res, err := c.cli.Indices.GetIndexTemplate(
		c.cli.Indices.GetIndexTemplate.WithName(templateName),
		c.cli.Indices.GetIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("检查索引模板失败: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, nil
	}

	if res.IsError() {
		return false, fmt.Errorf("检查索引模板响应错误: %s", res.String())
	}

	return true, nil
}

func buildMustQuery(conditions map[string]interface{}) []map[string]interface{} {
	var mustQuery []map[string]interface{}
	for field, value := range conditions {
		mustQuery = append(mustQuery, map[string]interface{}{
			"term": map[string]interface{}{
				field: value,
			},
		})
	}
	return mustQuery
}
