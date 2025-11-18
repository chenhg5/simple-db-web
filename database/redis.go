package database

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis 实现Database接口
// 注意：Redis 是键值存储数据库，需要特殊处理
type Redis struct {
	client  redis.UniversalClient // 使用 UniversalClient 支持单机和集群
	dbIndex int
	ctx     context.Context
}

// NewRedis 创建Redis实例
func NewRedis() *Redis {
	return &Redis{
		dbIndex: 0,
		ctx:     context.Background(),
	}
}

// Connect 建立Redis连接
func (r *Redis) Connect(dsn string) error {
	// Redis DSN格式支持多种：
	// 1. redis://:password@host:port/db (有密码)
	// 2. redis://host:port/db (无密码)
	// 3. host:port?password=xxx&db=0 (兼容格式)
	// 4. 直接 host:port (无密码，无数据库)

	var opts *redis.Options

	// 尝试解析 Redis URL 格式
	if strings.HasPrefix(dsn, "redis://") || strings.HasPrefix(dsn, "rediss://") {
		// 使用 redis URL 格式
		// 注意：redis.ParseURL 对于 redis://host:port/db 格式（无密码）能正确处理
		// 对于 redis://:password@host:port/db 格式（有密码）也能正确处理
		parsedOpts, err := redis.ParseURL(dsn)
		if err != nil {
			return fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		opts = parsedOpts
		// 确保 Password 字段正确（空字符串表示无密码）
		if opts.Password == "" {
			// 如果解析后的密码为空，确保不设置密码（redis.Options 的 Password 为空字符串时表示无密码）
		}
	} else {
		// 解析 host:port?password=xxx&db=0 格式
		opts = &redis.Options{}

		// 分离主机和参数
		parts := strings.Split(dsn, "?")
		addr := parts[0]

		// 解析地址
		if !strings.Contains(addr, ":") {
			addr = addr + ":6379" // 默认端口
		}
		opts.Addr = addr

		// 解析参数
		if len(parts) > 1 {
			params := strings.Split(parts[1], "&")
			for _, param := range params {
				kv := strings.Split(param, "=")
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					// URL 解码值（因为 BuildRedisDSN 中进行了 URL 编码）
					decodedValue, err := url.QueryUnescape(value)
					if err != nil {
						// 如果解码失败，使用原始值
						decodedValue = value
					}
					switch key {
					case "password":
						// 只有当值不为空时才设置密码（支持空密码）
						if decodedValue != "" {
							opts.Password = decodedValue
						}
					case "db":
						if db, err := strconv.Atoi(decodedValue); err == nil {
							opts.DB = db
							r.dbIndex = db
						}
					}
				}
			}
		}
	}

	// 确保 Password 字段正确设置（空字符串表示无密码，nil 也表示无密码）
	// redis.Options 的 Password 字段为空字符串时，表示不使用密码
	// 但如果 DSN 中没有 password 参数，opts.Password 应该保持为空字符串

	// 禁用客户端身份功能，避免在不支持 maint_notifications 的 Redis 版本上报错
	// 这个选项可以防止客户端尝试使用服务器不支持的命令
	// 注意：DisableIndentity 选项在 go-redis v9 中可用
	opts.DisableIndentity = true

	// 先尝试单机连接
	client := redis.NewClient(opts)

	// 测试连接并检测是否是集群
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// 尝试执行一个命令来检测是否是集群模式
	// 如果是集群，会返回 MOVED 或 ASK 错误
	// 使用 INFO cluster 命令来检测，更安全
	isCluster := false
	info, err := client.Info(ctx, "cluster").Result()
	if err == nil {
		// 如果 INFO cluster 命令成功，检查是否是集群模式
		if strings.Contains(info, "cluster_enabled:1") {
			isCluster = true
		}
	} else {
		// 如果 INFO 命令失败，尝试执行一个简单的命令
		// 如果返回 MOVED 或 ASK 错误，说明是集群
		testKey := "__cluster_test__"
		testErr := client.Set(ctx, testKey, "test", time.Second).Err()
		client.Del(ctx, testKey) // 清理测试键
		
		if testErr != nil {
			errStr := testErr.Error()
			if strings.Contains(errStr, "MOVED") || strings.Contains(errStr, "ASK") {
				isCluster = true
			}
		}
	}

	// 如果是集群模式，使用集群客户端
	if isCluster {
		client.Close() // 关闭单机客户端
		
		// 构建集群选项
		clusterOpts := &redis.ClusterOptions{
			Addrs:    []string{opts.Addr},
			Password: opts.Password,
		}
		clusterOpts.DisableIndentity = true
		
		clusterClient := redis.NewClusterClient(clusterOpts)
		
		// 测试集群连接
		if err := clusterClient.Ping(ctx).Err(); err != nil {
			clusterClient.Close()
			return fmt.Errorf("failed to connect to Redis cluster: %w", err)
		}
		
		r.client = clusterClient
	} else {
		r.client = client
		if opts != nil && opts.DB > 0 {
			r.dbIndex = opts.DB
		}
	}

	return nil
}

// Close 关闭连接
func (r *Redis) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// GetTypeName 获取数据库类型名称
func (r *Redis) GetTypeName() string {
	return "redis"
}

// GetDisplayName 获取数据库显示名称
func (r *Redis) GetDisplayName() string {
	return "Redis"
}

// GetTables 获取所有"表"（在Redis中，我们返回数据类型分类或键模式）
// 返回数据类型列表：strings, hashes, lists, sets, sorted_sets
func (r *Redis) GetTables() ([]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 返回 Redis 的数据类型分类（使用单数形式）
	return []string{"string", "hash", "list", "set", "zset", "keys"}, nil
}

// GetTableSchema 获取键的结构信息
func (r *Redis) GetTableSchema(tableName string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("database not connected")
	}

	// 如果 tableName 是数据类型，返回该类型的描述
	if tableName == "string" || tableName == "hash" || tableName == "list" ||
		tableName == "set" || tableName == "zset" || tableName == "keys" {
		return fmt.Sprintf("Redis data type: %s\nThis is a collection of keys of type %s", tableName, tableName), nil
	}

	// 如果 tableName 是一个具体的键，返回该键的详细信息
	keyType, err := r.client.Type(r.ctx, tableName).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key type: %w", err)
	}

	ttl, _ := r.client.TTL(r.ctx, tableName).Result()
	memory, _ := r.client.MemoryUsage(r.ctx, tableName).Result()

	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("Key: %s\n", tableName))
	schema.WriteString(fmt.Sprintf("Type: %s\n", keyType))
	if ttl > 0 {
		schema.WriteString(fmt.Sprintf("TTL: %s\n", ttl))
	} else if ttl == -1 {
		schema.WriteString("TTL: No expiration\n")
	}
	if memory > 0 {
		schema.WriteString(fmt.Sprintf("Memory: %d bytes\n", memory))
	}

	return schema.String(), nil
}

// GetTableColumns 获取表的列信息（Redis 中，我们根据数据类型返回不同的列）
func (r *Redis) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 根据数据类型返回不同的列信息
	// 注意：对于类型列表（如 "hash"），返回的是键列表的列，而不是该类型数据的列
	switch tableName {
	case "keys":
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	case "string":
		// 类型列表：显示键列表
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "value", Type: "string", Nullable: true},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	case "hash":
		// 类型列表：显示键列表
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "size", Type: "integer", Nullable: true},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	case "list":
		// 类型列表：显示键列表
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "size", Type: "integer", Nullable: true},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	case "set":
		// 类型列表：显示键列表
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "size", Type: "integer", Nullable: true},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	case "zset":
		// 类型列表：显示键列表
		return []ColumnInfo{
			{Name: "key", Type: "string", Nullable: false},
			{Name: "type", Type: "string", Nullable: false},
			{Name: "size", Type: "integer", Nullable: true},
			{Name: "ttl", Type: "integer", Nullable: true},
			{Name: "memory", Type: "integer", Nullable: true},
		}, nil
	default:
		// 尝试作为键名处理
		keyType, err := r.client.Type(r.ctx, tableName).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get key type: %w", err)
		}

		switch keyType {
		case "string":
			return []ColumnInfo{
				{Name: "key", Type: "string", Nullable: false},
				{Name: "value", Type: "string", Nullable: true},
			}, nil
		case "hash":
			return []ColumnInfo{
				{Name: "field", Type: "string", Nullable: false},
				{Name: "value", Type: "string", Nullable: true},
			}, nil
		case "list":
			return []ColumnInfo{
				{Name: "index", Type: "integer", Nullable: false},
				{Name: "value", Type: "string", Nullable: true},
			}, nil
		case "set":
			return []ColumnInfo{
				{Name: "member", Type: "string", Nullable: false},
			}, nil
		case "zset":
			return []ColumnInfo{
				{Name: "member", Type: "string", Nullable: false},
				{Name: "score", Type: "float", Nullable: false},
			}, nil
		default:
			return []ColumnInfo{
				{Name: "key", Type: "string", Nullable: false},
				{Name: "type", Type: "string", Nullable: false},
			}, nil
		}
	}
}

// GetTableData 获取表的数据（分页）
// tableName: 数据类型（strings, hashes等）或键名
func (r *Redis) GetTableData(tableName string, page, pageSize int, filters *FilterGroup) ([]map[string]interface{}, int64, error) {
	if r.client == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	// 如果 tableName 是数据类型，使用 SCAN 获取该类型的键
	if tableName == "keys" || tableName == "string" || tableName == "hash" ||
		tableName == "list" || tableName == "set" || tableName == "zset" {
		return r.getKeysByType(tableName, page, pageSize, filters)
	}

	// 否则，tableName 是一个具体的键，获取该键的数据
	return r.getKeyData(tableName, page, pageSize)
}

// getKeysByType 根据类型获取键列表（不支持完整分页，只获取当前页数据）
func (r *Redis) getKeysByType(dataType string, page, pageSize int, filters *FilterGroup) ([]map[string]interface{}, int64, error) {
	// 使用 SCAN 命令迭代获取键，只获取当前页需要的数据
	// 注意：Redis 不支持真正的分页，这里只获取当前页的数据，不扫描所有键

	pattern := "*"
	// 从过滤条件中提取模式（如果有）
	if filters != nil && len(filters.Conditions) > 0 {
		for _, condition := range filters.Conditions {
			if condition.Field == "key" && condition.Operator == "LIKE" && condition.Value != "" {
				// 将 SQL LIKE 模式转换为 Redis 模式
				pattern = strings.ReplaceAll(condition.Value, "%", "*")
				pattern = strings.ReplaceAll(pattern, "_", "?")
				break
			}
		}
	}

	// 计算需要跳过的键数量
	skipCount := (page - 1) * pageSize
	neededCount := pageSize
	scanTimes := 0
	maxScanTimes := 5

	var cursor uint64 = 0
	var matchedKeys []string

	// 使用 SCAN 迭代，直到获取到足够的匹配键
	for len(matchedKeys) < skipCount+neededCount && scanTimes < maxScanTimes {
		keys, nextCursor, err := r.client.Scan(r.ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to scan keys: %w", err)
		}
		scanTimes++

		// 如果指定了类型，过滤键
		if dataType != "keys" {
			for _, key := range keys {
				keyType, err := r.client.Type(r.ctx, key).Result()
				if err == nil {
					// 直接匹配 Redis 类型（使用单数形式）
					if keyType == "none" {
						continue // 跳过不存在的键
					}

					// 匹配类型（dataType 已经是单数形式，直接比较）
					if keyType == dataType {
						matchedKeys = append(matchedKeys, key)
					}
				}
			}
		} else {
			matchedKeys = append(matchedKeys, keys...)
		}

		if nextCursor == 0 {
			// 已经扫描完所有键
			break
		}
		cursor = nextCursor
	}

	// 如果匹配的键数量不足，返回空结果
	if len(matchedKeys) <= skipCount {
		return []map[string]interface{}{}, -1, nil // 返回 -1 表示总数未知
	}

	// 获取当前页的键
	start := skipCount
	end := min(start + neededCount, len(matchedKeys))

	keys := matchedKeys[start:end]

	// 构建结果
	results := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		keyType, _ := r.client.Type(r.ctx, key).Result()
		ttl, _ := r.client.TTL(r.ctx, key).Result()
		memory, _ := r.client.MemoryUsage(r.ctx, key).Result()

		row := map[string]interface{}{
			"key":  key,
			"type": keyType,
		}

		// 设置 TTL（即使为 -1 或 0 也设置为 nil）
		if ttl > 0 {
			row["ttl"] = ttl.Seconds()
		} else {
			row["ttl"] = nil
		}

		// 设置内存使用（即使为 0 也设置为 nil）
		if memory > 0 {
			row["memory"] = memory
		} else {
			row["memory"] = nil
		}

		// 根据类型添加预览值或大小
		switch keyType {
		case "string":
			// String 类型：显示值预览
			if val, err := r.client.Get(r.ctx, key).Result(); err == nil {
				// 限制预览长度
				if len(val) > 100 {
					row["value"] = val[:100] + "..."
				} else {
					row["value"] = val
				}
			} else {
				row["value"] = nil
			}
			// String 类型不需要 size 字段
			row["size"] = nil
		case "hash":
			// Hash 类型：显示大小
			if size, err := r.client.HLen(r.ctx, key).Result(); err == nil {
				row["size"] = size
			} else {
				row["size"] = nil
			}
			row["value"] = nil
		case "list":
			// List 类型：显示大小
			if size, err := r.client.LLen(r.ctx, key).Result(); err == nil {
				row["size"] = size
			} else {
				row["size"] = nil
			}
			row["value"] = nil
		case "set":
			// Set 类型：显示大小
			if size, err := r.client.SCard(r.ctx, key).Result(); err == nil {
				row["size"] = size
			} else {
				row["size"] = nil
			}
			row["value"] = nil
		case "zset":
			// Sorted Set 类型：显示大小
			if size, err := r.client.ZCard(r.ctx, key).Result(); err == nil {
				row["size"] = size
			} else {
				row["size"] = nil
			}
			row["value"] = nil
		default:
			// 其他类型：不设置 value 和 size
			row["value"] = nil
			row["size"] = nil
		}

		results = append(results, row)
	}

	// 返回 -1 表示总数未知（不支持完整分页）
	return results, -1, nil
}

// getKeyData 获取具体键的数据
func (r *Redis) getKeyData(key string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	keyType, err := r.client.Type(r.ctx, key).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get key type: %w", err)
	}

	switch keyType {
	case "string":
		val, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get key value: %w", err)
		}
		return []map[string]interface{}{
			{"key": key, "value": val},
		}, -1, nil

	case "hash":
		// 先检查键是否存在和类型是否正确
		exists, err := r.client.Exists(r.ctx, key).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to check key existence: %w", err)
		}
		if exists == 0 {
			return []map[string]interface{}{}, -1, fmt.Errorf("key does not exist")
		}

		// 获取 hash 的长度
		hashLen, err := r.client.HLen(r.ctx, key).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get hash length: %w", err)
		}

		// 如果 hash 为空，直接返回空结果
		if hashLen == 0 {
			return []map[string]interface{}{}, -1, nil
		}

		// 获取所有字段和值（Hash 类型需要获取全部数据才能分页）
		allFields, err := r.client.HGetAll(r.ctx, key).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get hash: %w", err)
		}

		// 检查是否获取到数据
		if len(allFields) == 0 {
			// 如果 HGetAll 返回空但 HLen 不为 0，可能是编码问题，尝试使用 HSCAN
			// 但为了简单起见，先返回空结果并记录警告
			return []map[string]interface{}{}, -1, nil
		}

		start := (page - 1) * pageSize
		end := start + pageSize

		// 将 map 转换为有序的切片以便分页
		// 使用切片保持字段顺序（虽然 map 本身无序，但我们可以按字段名排序）
		fieldList := make([]struct {
			field string
			value string
		}, 0, len(allFields))
		for field, value := range allFields {
			fieldList = append(fieldList, struct {
				field string
				value string
			}{field, value})
		}

		// 按字段名排序，确保结果顺序一致
		sort.Slice(fieldList, func(i, j int) bool {
			return fieldList[i].field < fieldList[j].field
		})

		if start >= len(fieldList) {
			return []map[string]interface{}{}, -1, nil
		}
		if end > len(fieldList) {
			end = len(fieldList)
		}

		results := make([]map[string]interface{}, 0, end-start)
		for i := start; i < end; i++ {
			results = append(results, map[string]interface{}{
				"field": fieldList[i].field,
				"value": fieldList[i].value,
			})
		}

		// 返回 -1 表示总数未知（不支持完整分页）
		return results, -1, nil

	case "list":
		// List 类型支持范围查询，但总数未知
		start := int64((page - 1) * pageSize)
		end := start + int64(pageSize) - 1

		values, err := r.client.LRange(r.ctx, key, start, end).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get list range: %w", err)
		}

		results := make([]map[string]interface{}, 0, len(values))
		for i, value := range values {
			results = append(results, map[string]interface{}{
				"index": start + int64(i),
				"value": value,
			})
		}

		// 返回 -1 表示总数未知（不支持完整分页）
		return results, -1, nil

	case "set":
		// 使用 SSCAN 迭代获取成员，只获取当前页需要的数据
		skipCount := (page - 1) * pageSize
		neededCount := pageSize

		var allMembers []string
		var cursor uint64 = 0

		// 迭代直到获取到足够的成员
		for len(allMembers) < skipCount+neededCount {
			members, nextCursor, err := r.client.SScan(r.ctx, key, cursor, "", int64(neededCount*2)).Result()
			if err != nil {
				return nil, -1, fmt.Errorf("failed to scan set: %w", err)
			}
			allMembers = append(allMembers, members...)
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}

		if len(allMembers) <= skipCount {
			return []map[string]interface{}{}, -1, nil
		}

		start := skipCount
		end := start + neededCount
		if end > len(allMembers) {
			end = len(allMembers)
		}

		members := allMembers[start:end]
		results := make([]map[string]interface{}, 0, len(members))
		for _, member := range members {
			results = append(results, map[string]interface{}{
				"member": member,
			})
		}

		// 返回 -1 表示总数未知（不支持完整分页）
		return results, -1, nil

	case "zset":
		// Sorted Set 类型支持范围查询，但总数未知
		start := int64((page - 1) * pageSize)
		end := start + int64(pageSize) - 1

		members, err := r.client.ZRangeWithScores(r.ctx, key, start, end).Result()
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get sorted set range: %w", err)
		}

		results := make([]map[string]interface{}, 0, len(members))
		for _, member := range members {
			results = append(results, map[string]interface{}{
				"member": member.Member,
				"score":  member.Score,
			})
		}

		// 返回 -1 表示总数未知（不支持完整分页）
		return results, -1, nil

	default:
		return nil, -1, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GetTableDataByID 基于主键ID获取表数据（Redis 不支持基于ID的分页，直接调用 GetTableData）
func (r *Redis) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string, filters *FilterGroup) ([]map[string]interface{}, int64, interface{}, error) {
	// Redis 不支持基于ID的分页，直接使用普通分页，页码固定为1
	data, total, err := r.GetTableData(tableName, 1, pageSize, filters)
	return data, total, nil, err
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（Redis 不支持）
func (r *Redis) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	// Redis 不支持基于ID的分页
	return nil, fmt.Errorf("Redis does not support ID-based pagination")
}

// ExecuteQuery 执行查询（Redis 命令）
// 支持的命令格式：
//   - GET key
//   - SET key value
//   - HGETALL key
//   - LRANGE key start stop
//   - SMEMBERS key
//   - ZRANGE key start stop
//   - SCAN cursor [MATCH pattern] [COUNT count]
//   - INFO [section]
func (r *Redis) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 解析命令
	parts := strings.Fields(strings.TrimSpace(query))
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	command := strings.ToUpper(parts[0])
	args := parts[1:]

	switch command {
	case "GET":
		if len(args) < 1 {
			return nil, fmt.Errorf("GET command requires a key")
		}
		val, err := r.client.Get(r.ctx, args[0]).Result()
		if err == redis.Nil {
			return []map[string]interface{}{{"key": args[0], "value": nil}}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to execute GET: %w", err)
		}
		return []map[string]interface{}{{"key": args[0], "value": val}}, nil

	case "HGETALL":
		if len(args) < 1 {
			return nil, fmt.Errorf("HGETALL command requires a key")
		}
		fields, err := r.client.HGetAll(r.ctx, args[0]).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute HGETALL: %w", err)
		}
		results := make([]map[string]interface{}, 0, len(fields))
		for field, value := range fields {
			results = append(results, map[string]interface{}{
				"field": field,
				"value": value,
			})
		}
		return results, nil

	case "LRANGE":
		if len(args) < 3 {
			return nil, fmt.Errorf("LRANGE command requires key start stop")
		}
		start, err1 := strconv.Atoi(args[1])
		stop, err2 := strconv.Atoi(args[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("LRANGE start and stop must be integers")
		}
		values, err := r.client.LRange(r.ctx, args[0], int64(start), int64(stop)).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute LRANGE: %w", err)
		}
		results := make([]map[string]interface{}, 0, len(values))
		for i, value := range values {
			results = append(results, map[string]interface{}{
				"index": start + i,
				"value": value,
			})
		}
		return results, nil

	case "SMEMBERS":
		if len(args) < 1 {
			return nil, fmt.Errorf("SMEMBERS command requires a key")
		}
		members, err := r.client.SMembers(r.ctx, args[0]).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute SMEMBERS: %w", err)
		}
		results := make([]map[string]interface{}, 0, len(members))
		for _, member := range members {
			results = append(results, map[string]interface{}{
				"member": member,
			})
		}
		return results, nil

	case "ZRANGE":
		if len(args) < 3 {
			return nil, fmt.Errorf("ZRANGE command requires key start stop")
		}
		start, err1 := strconv.Atoi(args[1])
		stop, err2 := strconv.Atoi(args[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("ZRANGE start and stop must be integers")
		}
		members, err := r.client.ZRangeWithScores(r.ctx, args[0], int64(start), int64(stop)).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute ZRANGE: %w", err)
		}
		results := make([]map[string]interface{}, 0, len(members))
		for _, member := range members {
			results = append(results, map[string]interface{}{
				"member": member.Member,
				"score":  member.Score,
			})
		}
		return results, nil

	case "INFO":
		section := ""
		if len(args) > 0 {
			section = args[0]
		}
		info, err := r.client.Info(r.ctx, section).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute INFO: %w", err)
		}
		// 解析 INFO 输出为键值对
		lines := strings.Split(info, "\n")
		results := make([]map[string]interface{}, 0)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				results = append(results, map[string]interface{}{
					"key":   strings.TrimSpace(parts[0]),
					"value": strings.TrimSpace(parts[1]),
				})
			}
		}
		return results, nil

	case "KEYS":
		if len(args) < 1 {
			return nil, fmt.Errorf("KEYS command requires a pattern")
		}
		keys, err := r.client.Keys(r.ctx, args[0]).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to execute KEYS: %w", err)
		}
		results := make([]map[string]interface{}, 0, len(keys))
		for _, key := range keys {
			results = append(results, map[string]interface{}{
				"key": key,
			})
		}
		return results, nil

	default:
		return nil, fmt.Errorf("unsupported Redis command: %s. Supported commands: GET, HGETALL, LRANGE, SMEMBERS, ZRANGE, INFO, KEYS", command)
	}
}

// ExecuteUpdate 执行更新（Redis SET, HSET 等命令）
func (r *Redis) ExecuteUpdate(query string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("database not connected")
	}

	parts := strings.Fields(strings.TrimSpace(query))
	if len(parts) < 2 {
		return 0, fmt.Errorf("update command requires at least command and key")
	}

	command := strings.ToUpper(parts[0])
	args := parts[1:]

	switch command {
	case "SET":
		if len(args) < 2 {
			return 0, fmt.Errorf("SET command requires key and value")
		}
		err := r.client.Set(r.ctx, args[0], strings.Join(args[1:], " "), 0).Err()
		if err != nil {
			return 0, fmt.Errorf("failed to execute SET: %w", err)
		}
		return 1, nil

	case "HSET":
		if len(args) < 3 {
			return 0, fmt.Errorf("HSET command requires key, field and value")
		}
		err := r.client.HSet(r.ctx, args[0], args[1], strings.Join(args[2:], " ")).Err()
		if err != nil {
			return 0, fmt.Errorf("failed to execute HSET: %w", err)
		}
		return 1, nil

	default:
		return 0, fmt.Errorf("unsupported update command: %s. Supported commands: SET, HSET", command)
	}
}

// ExecuteDelete 执行删除（Redis DEL 命令）
func (r *Redis) ExecuteDelete(query string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("database not connected")
	}

	parts := strings.Fields(strings.TrimSpace(query))
	if len(parts) < 2 {
		return 0, fmt.Errorf("delete command requires DEL and key(s)")
	}

	command := strings.ToUpper(parts[0])
	if command != "DEL" {
		return 0, fmt.Errorf("delete command must be DEL")
	}

	keys := parts[1:]
	count, err := r.client.Del(r.ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to execute DEL: %w", err)
	}

	return count, nil
}

// ExecuteInsert 执行插入（Redis SET, HSET 等命令，与 ExecuteUpdate 相同）
func (r *Redis) ExecuteInsert(query string) (int64, error) {
	return r.ExecuteUpdate(query)
}

// GetDatabases 获取所有数据库索引（Redis 默认有 16 个数据库，索引 0-15）
// 注意：集群模式只支持数据库 0
func (r *Redis) GetDatabases() ([]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// 检查是否是集群客户端（集群只支持数据库 0）
	_, isCluster := r.client.(*redis.ClusterClient)
	if isCluster {
		// 集群模式只支持数据库 0
		return []string{"0"}, nil
	}

	// Redis 默认有 16 个数据库（可以通过配置修改）
	// 返回所有可用的数据库索引
	databases := make([]string, 0, 16)
	for i := range 16 {
		databases = append(databases, strconv.Itoa(i))
	}

	return databases, nil
}

// SwitchDatabase 切换当前使用的数据库（Redis SELECT 命令）
// 注意：集群模式不支持 SELECT 命令
func (r *Redis) SwitchDatabase(databaseName string) error {
	if r.client == nil {
		return fmt.Errorf("database not connected")
	}

	// 检查是否是集群客户端（集群不支持 SELECT 命令）
	_, isCluster := r.client.(*redis.ClusterClient)
	if isCluster {
		return nil
	}

	dbIndex, err := strconv.Atoi(databaseName)
	if err != nil {
		return fmt.Errorf("invalid database index: %s", databaseName)
	}

	if err := r.client.Do(r.ctx, "SELECT", dbIndex).Err(); err != nil {
		return fmt.Errorf("failed to switch database: %w", err)
	}

	r.dbIndex = dbIndex
	return nil
}

// BuildRedisDSN 根据连接信息构建Redis DSN
func BuildRedisDSN(info ConnectionInfo) string {
	// 如果提供了 DSN，直接使用
	if info.DSN != "" {
		return info.DSN
	}

	// 构建 Redis DSN
	// 格式: redis://:password@host:port/db (有密码)
	// 或: redis://host:port/db (无密码)
	// 或: host:port?password=xxx&db=0 (兼容格式)

	host := info.Host
	if host == "" {
		host = "localhost"
	}

	port := info.Port
	if port == "" {
		port = "6379"
	}

	// 使用 host:port?password=xxx&db=0 格式（更灵活，支持空密码）
	dsn := fmt.Sprintf("%s:%s", host, port)
	
	params := []string{}
	
	// 如果有密码，添加密码参数（需要 URL 编码，防止特殊字符破坏 DSN 格式）
	if info.Password != "" {
		encodedPassword := url.QueryEscape(info.Password)
		params = append(params, "password="+encodedPassword)
	}
	
	// 如果有数据库索引，添加 db 参数
	if info.Database != "" {
		params = append(params, "db="+info.Database)
	}
	
	// 如果有参数，添加到 DSN
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
