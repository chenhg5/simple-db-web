package database

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
)

// Elasticsearch 实现Database接口
// 注意：Elasticsearch 是文档数据库，需要特殊处理
type Elasticsearch struct {
	client *elastic.Client
	ctx    context.Context
}

// NewElasticsearch 创建Elasticsearch实例
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		ctx: context.Background(),
	}
}

// Connect 建立Elasticsearch连接
func (e *Elasticsearch) Connect(dsn string) error {
	// Elasticsearch DSN格式支持多种：
	// 1. http://user:password@host:port (有认证)
	// 2. http://host:port (无认证)
	// 3. https://host:port (SSL)
	// 4. host:port?username=xxx&password=xxx&scheme=http (兼容格式)

	var scheme = "http"
	var host = "localhost"
	var port = "9200"
	var username = ""
	var password = ""

	// 解析 DSN
	if strings.HasPrefix(dsn, "http://") || strings.HasPrefix(dsn, "https://") {
		// URL 格式
		parsedURL, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("failed to parse Elasticsearch URL: %w", err)
		}

		scheme = parsedURL.Scheme
		host = parsedURL.Hostname()
		if parsedURL.Port() != "" {
			port = parsedURL.Port()
		} else {
			if scheme == "https" {
				port = "443"
			} else {
				port = "9200"
			}
		}

		if parsedURL.User != nil {
			username = parsedURL.User.Username()
			if pwd, ok := parsedURL.User.Password(); ok {
				password = pwd
			}
		}
	} else {
		// host:port?username=xxx&password=xxx&scheme=http 格式
		parts := strings.Split(dsn, "?")
		addr := parts[0]

		// 解析地址
		if strings.Contains(addr, ":") {
			addrParts := strings.Split(addr, ":")
			host = addrParts[0]
			port = addrParts[1]
		} else {
			host = addr
			port = "9200"
		}

		// 解析参数
		if len(parts) > 1 {
			params := strings.Split(parts[1], "&")
			for _, param := range params {
				kv := strings.Split(param, "=")
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					// URL 解码值
					decodedValue, err := url.QueryUnescape(value)
					if err != nil {
						decodedValue = value
					}
					switch key {
					case "username", "user":
						username = decodedValue
					case "password":
						password = decodedValue
					case "scheme":
						scheme = decodedValue
					}
				}
			}
		}
	}

	// 构建 Elasticsearch 客户端选项
	esURL := fmt.Sprintf("%s://%s:%s", scheme, host, port)
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(esURL),
		elastic.SetSniff(false),       // 禁用节点嗅探，避免连接问题
		elastic.SetHealthcheck(false), // 禁用自动健康检查，手动进行连接测试
		elastic.SetMaxRetries(0),      // 禁用重试，快速失败以便提供清晰的错误信息
	}

	// 如果是 HTTPS，配置跳过 SSL 证书验证（类似 curl --insecure）
	if scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过 SSL 证书验证
			},
		}
		httpClient := &http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		}
		options = append(options, elastic.SetHttpClient(httpClient))
	}

	// 如果有认证信息，添加
	if username != "" || password != "" {
		options = append(options, elastic.SetBasicAuth(username, password))
	}

	// 创建客户端
	client, err := elastic.NewClient(options...)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping 方法：如果传入 URL，会 ping 指定的 URL；如果不传，会使用客户端配置的 URL
	// 这里传入 URL 以确保 ping 正确的地址
	info, code, err := client.Ping(esURL).Do(ctx)
	if err != nil {
		client.Stop()
		// 提供更详细的错误信息，帮助用户诊断问题
		return fmt.Errorf("failed to connect to Elasticsearch at %s: %w (please check: 1) Elasticsearch is running, 2) network connectivity, 3) firewall settings, 4) Elasticsearch is listening on %s:%s)", esURL, err, host, port)
	}

	if code != 200 {
		client.Stop()
		return fmt.Errorf("Elasticsearch returned non-200 status: %d", code)
	}

	e.client = client
	_ = info // 可以记录版本信息

	return nil
}

// Close 关闭连接
func (e *Elasticsearch) Close() error {
	if e.client != nil {
		e.client.Stop()
	}
	return nil
}

// GetTypeName 获取数据库类型名称
func (e *Elasticsearch) GetTypeName() string {
	return "elasticsearch"
}

// GetDisplayName 获取数据库显示名称
func (e *Elasticsearch) GetDisplayName() string {
	return "Elasticsearch"
}

// GetTables 获取所有索引（在Elasticsearch中，索引相当于表）
func (e *Elasticsearch) GetTables() ([]string, error) {
	if e.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 获取所有索引
	indices, err := e.client.IndexNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get indices: %w", err)
	}

	// 过滤掉系统索引（以 . 开头的索引）
	var userIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			userIndices = append(userIndices, index)
		}
	}

	return userIndices, nil
}

// GetTableSchema 获取索引的结构信息（Mapping）
func (e *Elasticsearch) GetTableSchema(tableName string) (string, error) {
	if e.client == nil {
		return "", fmt.Errorf("database not connected")
	}

	// 先检查索引是否存在
	exists, err := e.client.IndexExists(tableName).Do(e.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check if index exists: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("index '%s' does not exist", tableName)
	}

	// 获取索引的 mapping
	// 尝试使用 PerformRequest 直接调用 API，避免客户端库可能添加的额外参数
	var mapping map[string]interface{}
	path := fmt.Sprintf("/%s/_mapping", tableName)
	res, err := e.client.PerformRequest(e.ctx, elastic.PerformRequestOptions{
		Method: "GET",
		Path:   path,
		Params: url.Values{}, // 不添加任何额外参数
	})
	if err != nil {
		// 如果直接调用失败，尝试使用标准方法
		mappingResult, err2 := e.client.GetMapping().Index(tableName).Do(e.ctx)
		if err2 != nil {
			return "", fmt.Errorf("failed to get mapping for index '%s': direct API call failed: %w, standard method also failed: %w", tableName, err, err2)
		}
		// 如果标准方法成功，使用标准方法的结果
		mapping = mappingResult
	} else {
		// 解析响应
		if err := json.Unmarshal(res.Body, &mapping); err != nil {
			return "", fmt.Errorf("failed to parse mapping response: %w", err)
		}
	}

	// 格式化输出
	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("Index: %s\n\n", tableName))

	// 获取索引的 settings
	settings, err := e.client.IndexGetSettings(tableName).Do(e.ctx)
	if err == nil && len(settings) > 0 {
		if indexSettings, ok := settings[tableName]; ok {
			schema.WriteString("Settings:\n")
			settingsJSON, _ := json.MarshalIndent(indexSettings.Settings, "", "  ")
			schema.WriteString(string(settingsJSON))
			schema.WriteString("\n\n")
		}
	}

	// 获取 mapping
	if len(mapping) > 0 {
		if indexMapping, ok := mapping[tableName]; ok {
			schema.WriteString("Mapping:\n")
			mappingJSON, _ := json.MarshalIndent(indexMapping, "", "  ")
			schema.WriteString(string(mappingJSON))
		} else {
			// 如果没有找到对应的索引 mapping，尝试直接使用返回的 mapping
			schema.WriteString("Mapping:\n")
			mappingJSON, _ := json.MarshalIndent(mapping, "", "  ")
			schema.WriteString(string(mappingJSON))
		}
	} else {
		schema.WriteString("Mapping: No mapping found\n")
	}

	return schema.String(), nil
}

// GetTableColumns 获取索引的列信息（从 Mapping 中提取字段）
func (e *Elasticsearch) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	if e.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 先检查索引是否存在
	exists, err := e.client.IndexExists(tableName).Do(e.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if index exists: %w", err)
	}
	if !exists {
		// 索引不存在，返回默认字段
		return []ColumnInfo{
			{Name: "_id", Type: "string", Nullable: false},
			{Name: "_source", Type: "object", Nullable: true},
		}, nil
	}

	// 获取索引的 mapping
	// 使用 PerformRequest 直接调用 API，避免客户端库可能添加的额外参数
	var mapping map[string]interface{}
	path := fmt.Sprintf("/%s/_mapping", tableName)
	res, err := e.client.PerformRequest(e.ctx, elastic.PerformRequestOptions{
		Method: "GET",
		Path:   path,
		Params: url.Values{}, // 不添加任何额外参数
	})
	if err != nil {
		// 如果直接调用失败，尝试使用标准方法
		mappingResult, err2 := e.client.GetMapping().Index(tableName).Do(e.ctx)
		if err2 != nil {
			// 如果两种方法都失败，返回默认字段而不是错误
			// 这样可以避免因为 mapping 获取失败而无法查看数据
			return []ColumnInfo{
				{Name: "_id", Type: "string", Nullable: false},
				{Name: "_source", Type: "object", Nullable: true},
			}, nil
		}
		mapping = mappingResult
	} else {
		// 解析响应
		if err := json.Unmarshal(res.Body, &mapping); err != nil {
			// 解析失败，返回默认字段
			return []ColumnInfo{
				{Name: "_id", Type: "string", Nullable: false},
				{Name: "_source", Type: "object", Nullable: true},
			}, nil
		}
	}

	var columns []ColumnInfo

	// 解析 mapping 提取字段
	if len(mapping) > 0 {
		// 尝试从 mapping 中获取索引的 mapping
		var indexMapping interface{}
		if im, ok := mapping[tableName]; ok {
			indexMapping = im
		} else {
			// 如果没有找到对应的索引，尝试使用第一个 mapping
			for _, v := range mapping {
				indexMapping = v
				break
			}
		}

		if indexMapping != nil {
			// 提取字段信息
			columns = extractFieldsFromMapping(indexMapping, "")
		}
	}

	// 如果没有找到字段，至少返回 _id 字段
	if len(columns) == 0 {
		columns = []ColumnInfo{
			{Name: "_id", Type: "string", Nullable: false},
			{Name: "_source", Type: "object", Nullable: true},
		}
	}

	return columns, nil
}

// extractFieldsFromMapping 从 mapping 中递归提取字段信息
func extractFieldsFromMapping(mapping interface{}, prefix string) []ColumnInfo {
	var columns []ColumnInfo

	mappingMap, ok := mapping.(map[string]interface{})
	if !ok {
		return columns
	}

	// 检查是否有 properties 字段（这是字段定义的地方）
	if properties, ok := mappingMap["properties"].(map[string]interface{}); ok {
		for fieldName, fieldDef := range properties {
			fullName := fieldName
			if prefix != "" {
				fullName = prefix + "." + fieldName
			}

			fieldMap, ok := fieldDef.(map[string]interface{})
			if !ok {
				continue
			}

			// 获取字段类型
			fieldType := "text"
			if t, ok := fieldMap["type"].(string); ok {
				fieldType = t
			}

			// 检查是否可空
			nullable := true
			if nullValue, ok := fieldMap["null_value"]; ok && nullValue != nil {
				nullable = false
			}

			columns = append(columns, ColumnInfo{
				Name:     fullName,
				Type:     fieldType,
				Nullable: nullable,
			})

			// 如果是 object 或 nested 类型，递归提取子字段
			if fieldType == "object" || fieldType == "nested" {
				if nestedProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
					nestedColumns := extractFieldsFromMapping(map[string]interface{}{
						"properties": nestedProps,
					}, fullName)
					columns = append(columns, nestedColumns...)
				}
			}
		}
	}

	return columns
}

// GetTableData 获取索引的数据（分页）
func (e *Elasticsearch) GetTableData(tableName string, page, pageSize int, filters *FilterGroup) ([]map[string]interface{}, int64, error) {
	if e.client == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	// 构建查询
	var query elastic.Query = elastic.NewMatchAllQuery()

	// 应用过滤条件
	if filters != nil && len(filters.Conditions) > 0 {
		var queries []elastic.Query
		for _, condition := range filters.Conditions {
			var q elastic.Query
			switch condition.Operator {
			case "=":
				q = elastic.NewTermQuery(condition.Field, condition.Value)
			case "!=":
				q = elastic.NewBoolQuery().MustNot(elastic.NewTermQuery(condition.Field, condition.Value))
			case "LIKE":
				q = elastic.NewWildcardQuery(condition.Field, "*"+condition.Value+"*")
			case ">":
				q = elastic.NewRangeQuery(condition.Field).Gt(condition.Value)
			case ">=":
				q = elastic.NewRangeQuery(condition.Field).Gte(condition.Value)
			case "<":
				q = elastic.NewRangeQuery(condition.Field).Lt(condition.Value)
			case "<=":
				q = elastic.NewRangeQuery(condition.Field).Lte(condition.Value)
			case "IS NULL":
				q = elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery(condition.Field))
			case "IS NOT NULL":
				q = elastic.NewExistsQuery(condition.Field)
			default:
				q = elastic.NewMatchAllQuery()
			}
			queries = append(queries, q)
		}

		// 组合查询
		if filters.Logic == "OR" {
			query = elastic.NewBoolQuery().Should(queries...).MinimumShouldMatch("1")
		} else {
			query = elastic.NewBoolQuery().Must(queries...)
		}
	}

	// 执行搜索
	from := (page - 1) * pageSize
	searchResult, err := e.client.Search().
		Index(tableName).
		Query(query).
		From(from).
		Size(pageSize).
		Do(e.ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to search: %w", err)
	}

	// 转换结果
	results := make([]map[string]interface{}, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		row := make(map[string]interface{})
		row["_id"] = hit.Id
		row["_score"] = hit.Score

		// 解析 _source
		if hit.Source != nil {
			var source map[string]interface{}
			if err := json.Unmarshal(hit.Source, &source); err == nil {
				// 将 _source 中的字段展开到 row 中
				for k, v := range source {
					row[k] = v
				}
				// 同时保留 _source 字段本身（作为 JSON 字符串，方便查看完整内容）
				sourceJSON, _ := json.Marshal(source)
				row["_source"] = string(sourceJSON)
			} else {
				// 如果解析失败，至少保留原始 _source 内容
				row["_source"] = string(hit.Source)
			}
		} else {
			// 如果没有 _source，设置为 null
			row["_source"] = nil
		}

		results = append(results, row)
	}

	// 返回总数
	total := searchResult.Hits.TotalHits.Value

	return results, total, nil
}

// GetTableDataByID 基于ID获取数据（Elasticsearch 不支持基于ID的分页，使用普通分页）
func (e *Elasticsearch) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string, filters *FilterGroup) ([]map[string]interface{}, int64, interface{}, error) {
	// Elasticsearch 不支持基于ID的分页，使用普通分页，页码固定为1
	data, total, err := e.GetTableData(tableName, 1, pageSize, filters)
	return data, total, nil, err
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（Elasticsearch 不支持）
func (e *Elasticsearch) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	return nil, fmt.Errorf("Elasticsearch does not support ID-based pagination")
}

// ExecuteQuery 执行查询（Elasticsearch DSL 查询）
func (e *Elasticsearch) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if e.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 尝试解析为 JSON DSL 查询
	var searchQuery map[string]interface{}
	if err := json.Unmarshal([]byte(query), &searchQuery); err != nil {
		return nil, fmt.Errorf("invalid Elasticsearch query DSL: %w", err)
	}

	// 执行搜索（需要指定索引，这里使用通配符查询所有索引）
	searchResult, err := e.client.Search().
		Index("*").
		Source(searchQuery).
		Do(e.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// 转换结果
	results := make([]map[string]interface{}, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		row := make(map[string]interface{})
		row["_index"] = hit.Index
		row["_id"] = hit.Id
		row["_score"] = hit.Score

		// 解析 _source
		if hit.Source != nil {
			var source map[string]interface{}
			if err := json.Unmarshal(hit.Source, &source); err == nil {
				// 将 _source 中的字段展开到 row 中
				for k, v := range source {
					row[k] = v
				}
				// 同时保留 _source 字段本身（作为 JSON 字符串，方便查看完整内容）
				sourceJSON, _ := json.Marshal(source)
				row["_source"] = string(sourceJSON)
			} else {
				// 如果解析失败，至少保留原始 _source 内容
				row["_source"] = string(hit.Source)
			}
		} else {
			// 如果没有 _source，设置为 null
			row["_source"] = nil
		}

		results = append(results, row)
	}

	return results, nil
}

// ExecuteUpdate 执行更新（Elasticsearch 使用 Update API）
func (e *Elasticsearch) ExecuteUpdate(query string) (int64, error) {
	if e.client == nil {
		return 0, fmt.Errorf("database not connected")
	}

	// 解析更新命令格式：UPDATE index/id {doc: {...}}
	// 简化处理：直接使用 JSON 格式
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(query), &updateData); err != nil {
		return 0, fmt.Errorf("invalid update format: %w", err)
	}

	// 需要指定索引和文档ID
	index, _ := updateData["_index"].(string)
	id, _ := updateData["_id"].(string)
	doc, _ := updateData["doc"].(map[string]interface{})

	if index == "" || id == "" {
		return 0, fmt.Errorf("update requires _index and _id")
	}

	_, err := e.client.Update().
		Index(index).
		Id(id).
		Doc(doc).
		Do(e.ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to update: %w", err)
	}

	return 1, nil
}

// ExecuteDelete 执行删除（Elasticsearch 使用 Delete API）
func (e *Elasticsearch) ExecuteDelete(query string) (int64, error) {
	if e.client == nil {
		return 0, fmt.Errorf("database not connected")
	}

	// 解析删除命令格式：DELETE index/id 或 JSON 格式
	var deleteData map[string]interface{}
	if err := json.Unmarshal([]byte(query), &deleteData); err != nil {
		// 如果不是 JSON，尝试解析为简单格式：index/id
		parts := strings.Fields(query)
		if len(parts) >= 3 && strings.ToUpper(parts[0]) == "DELETE" {
			index := parts[1]
			id := parts[2]
			_, err := e.client.Delete().
				Index(index).
				Id(id).
				Do(e.ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to delete: %w", err)
			}
			return 1, nil
		}
		return 0, fmt.Errorf("invalid delete format")
	}

	index, _ := deleteData["_index"].(string)
	id, _ := deleteData["_id"].(string)

	if index == "" || id == "" {
		return 0, fmt.Errorf("delete requires _index and _id")
	}

	_, err := e.client.Delete().
		Index(index).
		Id(id).
		Do(e.ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete: %w", err)
	}

	return 1, nil
}

// ExecuteInsert 执行插入（Elasticsearch 使用 Index API）
func (e *Elasticsearch) ExecuteInsert(query string) (int64, error) {
	if e.client == nil {
		return 0, fmt.Errorf("database not connected")
	}

	// 解析插入命令格式：JSON 格式
	var insertData map[string]interface{}
	if err := json.Unmarshal([]byte(query), &insertData); err != nil {
		return 0, fmt.Errorf("invalid insert format: %w", err)
	}

	// 需要指定索引
	index, _ := insertData["_index"].(string)
	id, _ := insertData["_id"].(string)

	if index == "" {
		return 0, fmt.Errorf("insert requires _index")
	}

	// 移除 _index 和 _id，保留文档内容
	doc := make(map[string]interface{})
	for k, v := range insertData {
		if k != "_index" && k != "_id" {
			doc[k] = v
		}
	}

	indexService := e.client.Index().Index(index).BodyJson(doc)
	if id != "" {
		indexService = indexService.Id(id)
	}

	_, err := indexService.Do(e.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to insert: %w", err)
	}

	return 1, nil
}

// GetDatabases 获取所有数据库（Elasticsearch 不支持多数据库概念，返回空列表）
func (e *Elasticsearch) GetDatabases() ([]string, error) {
	// Elasticsearch 不支持多数据库概念
	return []string{"default"}, nil
}

// SwitchDatabase 切换数据库（Elasticsearch 不支持）
func (e *Elasticsearch) SwitchDatabase(databaseName string) error {
	return nil
}

// BuildElasticsearchDSN 根据连接信息构建Elasticsearch DSN
func BuildElasticsearchDSN(info ConnectionInfo) string {
	// 如果提供了 DSN，直接使用
	if info.DSN != "" {
		return info.DSN
	}

	// 构建 Elasticsearch DSN
	// 格式: http://user:password@host:port 或 https://host:port
	// 或: host:port?username=xxx&password=xxx&scheme=http

	scheme := "http"
	host := info.Host
	if host == "" {
		host = "localhost"
	}

	port := info.Port
	if port == "" {
		port = "9200"
	}

	// 如果有用户名或密码，使用 URL 格式
	if info.User != "" || info.Password != "" {
		// URL 编码密码
		encodedPassword := url.QueryEscape(info.Password)
		return fmt.Sprintf("%s://%s:%s@%s:%s", scheme, info.User, encodedPassword, host, port)
	}

	// 否则使用 host:port?scheme=http 格式
	dsn := fmt.Sprintf("%s:%s", host, port)
	params := []string{"scheme=" + scheme}

	// 如果有用户名，添加
	if info.User != "" {
		params = append(params, "username="+url.QueryEscape(info.User))
	}

	// 如果有密码，添加
	if info.Password != "" {
		params = append(params, "password="+url.QueryEscape(info.Password))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

