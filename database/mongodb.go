package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 实现Database接口
// 注意：MongoDB 是 NoSQL 数据库，需要特殊处理
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	ctx      context.Context
}

// NewMongoDB 创建MongoDB实例
func NewMongoDB() *MongoDB {
	return &MongoDB{
		ctx: context.Background(),
	}
}

// Connect 建立MongoDB连接
func (m *MongoDB) Connect(dsn string) error {
	// MongoDB DSN格式: mongodb://user:password@host:port/database
	clientOptions := options.Client().ApplyURI(dsn)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	m.client = client
	
	// 从 DSN 中提取数据库名
	if strings.Contains(dsn, "/") && !strings.HasSuffix(dsn, "/") {
		parts := strings.Split(dsn, "/")
		if len(parts) > 0 {
			dbName := parts[len(parts)-1]
			if strings.Contains(dbName, "?") {
				dbName = strings.Split(dbName, "?")[0]
			}
			if dbName != "" {
				m.database = client.Database(dbName)
			}
		}
	}

	return nil
}

// Close 关闭连接
func (m *MongoDB) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}

// GetTables 获取所有集合名（MongoDB 中表称为集合）
func (m *MongoDB) GetTables() ([]string, error) {
	if m.client == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return nil, fmt.Errorf("database not selected")
	}

	collections, err := m.database.ListCollectionNames(m.ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query collection list: %w", err)
	}

	return collections, nil
}

// GetTableSchema 获取集合结构（MongoDB 是文档数据库，没有固定结构）
func (m *MongoDB) GetTableSchema(tableName string) (string, error) {
	if m.client == nil {
		return "", fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return "", fmt.Errorf("database not selected")
	}

	collection := m.database.Collection(tableName)
	
	// 获取一个示例文档来推断结构
	var sampleDoc bson.M
	err := collection.FindOne(m.ctx, bson.M{}).Decode(&sampleDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Sprintf("集合 %s 为空，无法推断结构", tableName), nil
		}
		return "", fmt.Errorf("failed to query collection schema: %w", err)
	}

	// 构建结构描述
	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("Collection: %s\n", tableName))
	schema.WriteString("Document structure (example):\n")
	
	for key, value := range sampleDoc {
		schema.WriteString(fmt.Sprintf("  %s: %T\n", key, value))
	}

	return schema.String(), nil
}

// GetTableColumns 获取集合的字段信息（MongoDB 是文档数据库，字段可能不固定）
func (m *MongoDB) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	if m.client == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return nil, fmt.Errorf("database not selected")
	}

	collection := m.database.Collection(tableName)
	
	// 获取一个示例文档来推断字段
	var sampleDoc bson.M
	err := collection.FindOne(m.ctx, bson.M{}).Decode(&sampleDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []ColumnInfo{}, nil
		}
		return nil, fmt.Errorf("failed to query collection fields: %w", err)
	}

	var columns []ColumnInfo
	for key, value := range sampleDoc {
		col := ColumnInfo{
			Name:     key,
			Type:     fmt.Sprintf("%T", value),
			Nullable: true, // MongoDB 字段都是可选的
			Key:      "",  // MongoDB 使用 _id 作为主键
		}
		if key == "_id" {
			col.Key = "PRI"
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// ExecuteQuery 执行查询（MongoDB 使用 JSON 查询）
func (m *MongoDB) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return nil, fmt.Errorf("database not selected")
	}

	// MongoDB 查询需要解析为 JSON/BSON
	// 这里简化处理，假设查询格式为: db.collection.find({...})
	// 实际使用时需要更复杂的解析逻辑
	
	// 尝试解析查询字符串
	// 格式: db.collectionName.find({...}) 或 collectionName.find({...})
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return nil, fmt.Errorf("MongoDB query format error, expected: collectionName.find({...})")
	}

	collectionName := parts[0]
	if strings.HasPrefix(collectionName, "db.") {
		collectionName = strings.TrimPrefix(collectionName, "db.")
	}

	collection := m.database.Collection(collectionName)
	
	// 解析查询条件（简化处理）
	var filter bson.M
	if len(parts) > 2 {
		// 尝试解析 JSON 查询条件
		filterStr := strings.Join(parts[2:], " ")
		if err := bson.UnmarshalExtJSON([]byte(filterStr), true, &filter); err != nil {
			// 如果解析失败，使用空查询
			filter = bson.M{}
		}
	} else {
		filter = bson.M{}
	}

	cursor, err := collection.Find(m.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close(m.ctx)

	var results = make([]map[string]interface{}, 0)
	for cursor.Next(m.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		// 转换 ObjectID 为字符串
		result := make(map[string]interface{})
		for key, value := range doc {
			if oid, ok := value.(primitive.ObjectID); ok {
				result[key] = oid.Hex()
			} else {
				result[key] = value
			}
		}
		results = append(results, result)
	}

	return results, cursor.Err()
}

// ExecuteUpdate 执行更新
func (m *MongoDB) ExecuteUpdate(query string) (int64, error) {
	if m.client == nil {
		return 0, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return 0, fmt.Errorf("database not selected")
	}

	// MongoDB 更新需要解析为 update 命令
	// 格式: collectionName.update({filter}, {update})
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return 0, fmt.Errorf("MongoDB update format error")
	}

	collectionName := parts[0]
	if strings.HasPrefix(collectionName, "db.") {
		collectionName = strings.TrimPrefix(collectionName, "db.")
	}

	// 简化处理：这里需要更复杂的解析逻辑
	// 暂时返回错误，提示使用正确的 MongoDB 更新语法
	_ = collectionName // 避免未使用变量错误
	return 0, fmt.Errorf("MongoDB update operation requires special syntax parsing, please use MongoDB native syntax")
}

// ExecuteDelete 执行删除
func (m *MongoDB) ExecuteDelete(query string) (int64, error) {
	if m.client == nil {
		return 0, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return 0, fmt.Errorf("database not selected")
	}

	// MongoDB 删除需要解析为 delete 命令
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return 0, fmt.Errorf("MongoDB delete format error")
	}

	collectionName := parts[0]
	if strings.HasPrefix(collectionName, "db.") {
		collectionName = strings.TrimPrefix(collectionName, "db.")
	}

	collection := m.database.Collection(collectionName)
	
	// 解析删除条件
	var filter bson.M
	if len(parts) > 2 {
		filterStr := strings.Join(parts[2:], " ")
		if err := bson.UnmarshalExtJSON([]byte(filterStr), true, &filter); err != nil {
			filter = bson.M{}
		}
	} else {
		filter = bson.M{}
	}

	result, err := collection.DeleteMany(m.ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %w", err)
	}

	return result.DeletedCount, nil
}

// ExecuteInsert 执行插入
func (m *MongoDB) ExecuteInsert(query string) (int64, error) {
	if m.client == nil {
		return 0, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return 0, fmt.Errorf("database not selected")
	}

	// MongoDB 插入需要解析为 insert 命令
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return 0, fmt.Errorf("MongoDB insert format error")
	}

	collectionName := parts[0]
	if strings.HasPrefix(collectionName, "db.") {
		collectionName = strings.TrimPrefix(collectionName, "db.")
	}

	collection := m.database.Collection(collectionName)
	
	// 解析插入文档
	var doc bson.M
	if len(parts) > 2 {
		docStr := strings.Join(parts[2:], " ")
		if err := bson.UnmarshalExtJSON([]byte(docStr), true, &doc); err != nil {
			return 0, fmt.Errorf("failed to parse insert document: %w", err)
		}
	} else {
		return 0, fmt.Errorf("missing insert document")
	}

	_, err := collection.InsertOne(m.ctx, doc)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}

	return 1, nil
}

// GetTableData 获取集合数据（分页）
func (m *MongoDB) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if m.client == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return nil, 0, fmt.Errorf("database not selected")
	}

	collection := m.database.Collection(tableName)

	// 获取总数
	total, err := collection.CountDocuments(m.ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query total count: %w", err)
	}

	// 获取分页数据
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := collection.Find(m.ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query data: %w", err)
	}
	defer cursor.Close(m.ctx)

	var results = make([]map[string]interface{}, 0)
	for cursor.Next(m.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}

		result := make(map[string]interface{})
		for key, value := range doc {
			if oid, ok := value.(primitive.ObjectID); ok {
				result[key] = oid.Hex()
			} else {
				result[key] = value
			}
		}
		results = append(results, result)
	}

	return results, total, cursor.Err()
}

// GetTableDataByID 基于主键ID获取表数据（高性能分页）
func (m *MongoDB) GetTableDataByID(tableName string, primaryKey string, lastId interface{}, pageSize int, direction string) ([]map[string]interface{}, int64, interface{}, error) {
	if m.client == nil {
		return nil, 0, nil, fmt.Errorf("数据库未连接")
	}
	if m.database == nil {
		return nil, 0, nil, fmt.Errorf("未选择数据库")
	}

	collection := m.database.Collection(tableName)

	// 获取总数
	total, err := collection.CountDocuments(m.ctx, bson.M{})
	if err != nil {
		return nil, 0, nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// MongoDB 使用 _id 作为主键
	if primaryKey == "" {
		primaryKey = "_id"
	}

	// 构建查询条件
	var filter bson.M
	var sort bson.M

	if direction == "prev" {
		if lastId == nil {
			return nil, 0, nil, fmt.Errorf("lastId is required for previous page")
		}
		// 转换 lastId 为 ObjectID（如果是字符串）
		var oid primitive.ObjectID
		if idStr, ok := lastId.(string); ok {
			oid, _ = primitive.ObjectIDFromHex(idStr)
		}
		filter = bson.M{primaryKey: bson.M{"$lt": oid}}
		sort = bson.M{primaryKey: -1}
	} else {
		if lastId != nil {
			var oid primitive.ObjectID
			if idStr, ok := lastId.(string); ok {
				oid, _ = primitive.ObjectIDFromHex(idStr)
			}
			filter = bson.M{primaryKey: bson.M{"$gt": oid}}
		} else {
			filter = bson.M{}
		}
		sort = bson.M{primaryKey: 1}
	}

	opts := options.Find().SetSort(sort).SetLimit(int64(pageSize))
	cursor, err := collection.Find(m.ctx, filter, opts)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer cursor.Close(m.ctx)

	var results = make([]map[string]interface{}, 0)
	var nextId interface{} = nil
	var firstId interface{} = nil

	for cursor.Next(m.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, nil, err
		}

		result := make(map[string]interface{})
		for key, value := range doc {
			if oid, ok := value.(primitive.ObjectID); ok {
				result[key] = oid.Hex()
				if key == primaryKey {
					if firstId == nil {
						firstId = oid.Hex()
					}
					nextId = oid.Hex()
				}
			} else {
				result[key] = value
				if key == primaryKey {
					if firstId == nil {
						firstId = value
					}
					nextId = value
				}
			}
		}
		results = append(results, result)
	}

	if direction == "prev" {
		// 反转结果
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
		nextId = firstId
	}

	return results, total, nextId, cursor.Err()
}

// GetPageIdByPageNumber 根据页码计算该页的起始ID（用于页码跳转）
func (m *MongoDB) GetPageIdByPageNumber(tableName string, primaryKey string, page, pageSize int) (interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if m.database == nil {
		return nil, fmt.Errorf("database not selected")
	}

	if page <= 1 {
		return nil, nil
	}

	collection := m.database.Collection(tableName)

	if primaryKey == "" {
		primaryKey = "_id"
	}

	skip := int64((page - 1) * pageSize - 1)
	opts := options.FindOne().SetSort(bson.M{primaryKey: 1}).SetSkip(skip)
	
	var doc bson.M
	err := collection.FindOne(m.ctx, bson.M{}, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query page ID: %w", err)
	}

	if idVal, ok := doc[primaryKey]; ok {
		if oid, ok := idVal.(primitive.ObjectID); ok {
			return oid.Hex(), nil
		}
		return idVal, nil
	}

	return nil, nil
}

// GetDatabases 获取所有数据库名称
func (m *MongoDB) GetDatabases() ([]string, error) {
	if m.client == nil {
		return nil, fmt.Errorf("database not connected")
	}

	databases, err := m.client.ListDatabaseNames(m.ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query database list: %w", err)
	}

	// 过滤系统数据库
	var userDatabases []string
	for _, db := range databases {
		if db != "admin" && db != "local" && db != "config" {
			userDatabases = append(userDatabases, db)
		}
	}

	return userDatabases, nil
}

// SwitchDatabase 切换当前使用的数据库
func (m *MongoDB) SwitchDatabase(databaseName string) error {
	if m.client == nil {
		return fmt.Errorf("database not connected")
	}
	m.database = m.client.Database(databaseName)
	return nil
}

// BuildMongoDBDSN 根据连接信息构建MongoDB DSN
func BuildMongoDBDSN(info ConnectionInfo) string {
	if info.DSN != "" {
		return info.DSN
	}

	// MongoDB DSN格式: mongodb://user:password@host:port/database
	var dsn string
	if info.User != "" && info.Password != "" {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%s",
			info.User,
			info.Password,
			info.Host,
			info.Port,
		)
	} else {
		dsn = fmt.Sprintf("mongodb://%s:%s",
			info.Host,
			info.Port,
		)
	}

	if info.Database != "" {
		dsn += "/" + info.Database
	}

	return dsn
}

