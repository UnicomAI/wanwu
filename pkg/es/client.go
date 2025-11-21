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
	// Intelligently determine the protocol. If the address does not have a protocol prefix, HTTPS will be tried. If it fails, HTTP will be tried.
	addresses := []string{}

	// If the address already contains a protocol, use it directly
	if strings.HasPrefix(c.Address, "http://") || strings.HasPrefix(c.Address, "https://") {
		addresses = append(addresses, c.Address)
	} else {
		// Try HTTPS first, then HTTP
		addresses = append(addresses, "https://"+c.Address, "http://"+c.Address)
	}

	var lastErr error

	// try every address
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
			lastErr = fmt.Errorf("failed to create ES client [%s]: %v", addr, err)
			log.Warnf("Failed to create ES client, address: %s, error: %v", addr, err)
			continue
		}

		// test connection
		res, err := esClient.Info()
		if err != nil {
			lastErr = fmt.Errorf("ES connection test failed [%s]: %v", addr, err)
			log.Warnf("ES connection test failed, address: %s, error: %v", addr, err)
			continue
		}

		if res != nil {
			defer res.Body.Close()

			if res.IsError() {
				lastErr = fmt.Errorf("ES connection response error [%s]: %s", addr, res.String())
				log.Warnf("ES connection response error, address: %s, response: %s", addr, res.String())
				continue
			}
		}

		log.Infof("ES connection succeeded, address: %s", addr)
		return &client{
			ctx:  ctx,
			cli:  esClient,
			stop: make(chan struct{}, 1),
		}, nil
	}

	// All addresses failed
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("failed to connect to ES, attempted addresses: %v", addresses)
}

func (c *client) Stop() {
	c.mutex.Lock()
	if c.stopped {
		log.Errorf("ES client already stopped")
		c.mutex.Unlock()
		return
	}
	c.stopped = true
	close(c.stop)
	c.mutex.Unlock()
	log.Infof("ES client stopped")
}

func (c *client) Cli() *elasticsearch.Client {
	return c.cli
}

// Write data to the specified index
func (c *client) IndexDocument(ctx context.Context, index string, document interface{}) error {
	docJSON, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to serialize document: %v", err)
	}

	res, err := c.cli.Index(
		index,
		strings.NewReader(string(docJSON)),
		c.cli.Index.WithContext(ctx),
		c.cli.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to write to ES: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES write response error: %s", res.String())
	}

	log.Infof("Successfully wrote to ES, index: %s", index)
	return nil
}

// Query all data based on specified field conditions
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
		return nil, 0, fmt.Errorf("failed to serialize query: %v", err)
	}

	res, err := c.cli.Search(
		c.cli.Search.WithContext(ctx),
		c.cli.Search.WithIndex(index),
		c.cli.Search.WithBody(strings.NewReader(string(queryJSON))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("ES query failed: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("ES query response error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse query result: %v", err)
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid query result format")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid total format")
	}

	totalValue, ok := total["value"].(float64)
	if !ok {
		return nil, 0, fmt.Errorf("invalid total value")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid hit list format")
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

	log.Infof("ES query succeeded, index: %s, total: %d, returned: %d", index, int64(totalValue), len(documents))
	return documents, int64(totalValue), nil
}

// Create index template
func (c *client) CreateIndexTemplate(ctx context.Context, templateName string, templateBody string) error {
	res, err := c.cli.Indices.PutIndexTemplate(
		templateName,
		strings.NewReader(templateBody),
		c.cli.Indices.PutIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create index template: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index template response error: %s", res.String())
	}

	log.Infof("Successfully created index template: %s", templateName)
	return nil
}

// Check if index template exists
func (c *client) IndexTemplateExists(ctx context.Context, templateName string) (bool, error) {
	res, err := c.cli.Indices.GetIndexTemplate(
		c.cli.Indices.GetIndexTemplate.WithName(templateName),
		c.cli.Indices.GetIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check index template: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, nil
	}

	if res.IsError() {
		return false, fmt.Errorf("check index template response error: %s", res.String())
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
