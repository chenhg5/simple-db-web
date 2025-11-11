package database

import (
	"context"
	"testing"

	"ksogit.kingsoft.net/chat/lib/xmysql"
	xmysqlv2 "ksogit.kingsoft.net/chat/lib/xmysql/v2"
)

func TestDamengDialect_getColumnsFromSchema(t *testing.T) {
	// 创建一个 DamengDialect 实例用于测试（不需要真实的数据库连接）
	dialect := &DamengDialect{BaseDialect: &BaseDialect{}}

	tests := []struct {
		name          string
		schema        string
		expectedCols  []ColumnInfo
		expectedError bool
	}{
		{
			name: "测试用例：kim_partition_dbinst 表",
			schema: `CREATE TABLE "test_kim_10132903552522188748"."kim_partition_dbinst"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"src_database" VARCHAR(64 CHAR) DEFAULT '' NOT NULL,
"src_table" VARCHAR(64 CHAR) DEFAULT '' NOT NULL,
"mgr_id" BIGINT DEFAULT 0 NOT NULL,
"db_inst" VARCHAR(64 CHAR) DEFAULT '' NOT NULL,
"dst_table" VARCHAR(64 CHAR) DEFAULT '' NOT NULL,
"start_pos" BIGINT DEFAULT 0 NOT NULL,
"end_pos" BIGINT DEFAULT 0 NOT NULL,
NOT CLUSTER PRIMARY KEY("id")) STORAGE(ON "MAIN", CLUSTERBTR) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "src_database", Type: "VARCHAR(64 CHAR)", DefaultValue: "", Key: ""},
				{Name: "src_table", Type: "VARCHAR(64 CHAR)", DefaultValue: "", Key: ""},
				{Name: "mgr_id", Type: "BIGINT", DefaultValue: "0", Key: ""},
				{Name: "db_inst", Type: "VARCHAR(64 CHAR)", DefaultValue: "", Key: ""},
				{Name: "dst_table", Type: "VARCHAR(64 CHAR)", DefaultValue: "", Key: ""},
				{Name: "start_pos", Type: "BIGINT", DefaultValue: "0", Key: ""},
				{Name: "end_pos", Type: "BIGINT", DefaultValue: "0", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试包含字符串 DEFAULT 值的列",
			schema: `CREATE TABLE "test_table"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"name" VARCHAR(100) DEFAULT 'test' NOT NULL,
"email" VARCHAR(255) DEFAULT '' NOT NULL,
"description" VARCHAR(500) DEFAULT 'default description' NOT NULL
) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "name", Type: "VARCHAR(100)", DefaultValue: "test", Key: ""},
				{Name: "email", Type: "VARCHAR(255)", DefaultValue: "", Key: ""},
				{Name: "description", Type: "VARCHAR(500)", DefaultValue: "default description", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试数字 DEFAULT 值（带单引号）",
			schema: `CREATE TABLE "test_table"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"age" INT DEFAULT '18' NOT NULL,
"score" DECIMAL(10,2) DEFAULT '0.00' NOT NULL
) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "age", Type: "INT", DefaultValue: "18", Key: ""},
				{Name: "score", Type: "DECIMAL(10,2)", DefaultValue: "0.00", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试空表结构",
			schema: `CREATE TABLE "empty_table"
(
) ;`,
			expectedCols: []ColumnInfo{},
			expectedError: false,
		},
		{
			name: "测试没有 DEFAULT 值的列",
			schema: `CREATE TABLE "test_table"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"name" VARCHAR(100) NOT NULL,
"created_at" TIMESTAMP NOT NULL
) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "name", Type: "VARCHAR(100)", DefaultValue: "", Key: ""},
				{Name: "created_at", Type: "TIMESTAMP", DefaultValue: "", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试混合情况：有 DEFAULT 和没有 DEFAULT 的列",
			schema: `CREATE TABLE "test_table"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"name" VARCHAR(100) NOT NULL,
"status" VARCHAR(20) DEFAULT 'active' NOT NULL,
"created_at" TIMESTAMP NOT NULL,
"updated_at" TIMESTAMP DEFAULT '2024-01-01' NOT NULL
) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "name", Type: "VARCHAR(100)", DefaultValue: "", Key: ""},
				{Name: "status", Type: "VARCHAR(20)", DefaultValue: "active", Key: ""},
				{Name: "created_at", Type: "TIMESTAMP", DefaultValue: "", Key: ""},
				{Name: "updated_at", Type: "TIMESTAMP", DefaultValue: "2024-01-01", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试复杂类型定义",
			schema: `CREATE TABLE "test_table"
(
"id" BIGINT IDENTITY(1, 1) NOT NULL,
"data" VARCHAR(100 CHAR) DEFAULT 'default' NOT NULL,
"value" NUMERIC(10,2) DEFAULT '0.00' NOT NULL
) ;`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "BIGINT", DefaultValue: "", Key: "PRI"},
				{Name: "data", Type: "VARCHAR(100 CHAR)", DefaultValue: "default", Key: ""},
				{Name: "value", Type: "NUMERIC(10,2)", DefaultValue: "0.00", Key: ""},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接调用私有方法 getColumnsFromSchema
			columns, err := dialect.getColumnsFromSchema(tt.schema)

			// 验证错误
			if tt.expectedError {
				if err == nil {
					t.Errorf("期望返回错误，但实际没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望返回错误，但实际返回: %v", err)
				return
			}

			// 验证列数量
			if len(columns) != len(tt.expectedCols) {
				t.Errorf("列数量不匹配: 期望 %d, 实际 %d", len(tt.expectedCols), len(columns))
				t.Logf("期望的列: %+v", tt.expectedCols)
				t.Logf("实际的列: %+v", columns)
				return
			}

			// 验证每列的信息
			for i, expectedCol := range tt.expectedCols {
				if i >= len(columns) {
					t.Errorf("列索引 %d 超出范围", i)
					continue
				}

				actualCol := columns[i]
				if actualCol.Name != expectedCol.Name {
					t.Errorf("列 %d 名称不匹配: 期望 %s, 实际 %s", i, expectedCol.Name, actualCol.Name)
				}
				if actualCol.Type != expectedCol.Type {
					t.Errorf("列 %d 类型不匹配: 期望 %s, 实际 %s", i, expectedCol.Type, actualCol.Type)
				}
				if actualCol.DefaultValue != expectedCol.DefaultValue {
					t.Errorf("列 %d 默认值不匹配: 期望 %s, 实际 %s", i, expectedCol.DefaultValue, actualCol.DefaultValue)
				}
			}
		})
	}
}

func TestQueryMetaData(t *testing.T) {
	// host := "10.13.83.245"
	// user := "vastwps"
	// password := "r9S3LjQ@C5FkW$"
	// port := 20142
	dbname := "kim"

	// KING_DB_HOST 10.13.83.244
	// KING_DB_PORT 20142
	// KING_DB_USER TEST01
	// KING_DB_PASSWD Vh3auMiCSBRdy@*-
	// host := "10.13.83.244"
	// user := "TEST01"
	// password := "Vh3auMiCSBRdy@*-"
	// port := 20142

	// OCEAN_DB_HOST 10.13.83.246
	// OCEAN_DB_PORT 20142
	// OCEAN_DB_USER root@mysqlt#obcluster
	// OCEAN_DB_PASSWD r9S3LjQ@C5FkW$
	host := "10.13.83.246"
	user := "root@mysqlt#obcluster"
	password := "r9S3LjQ@C5FkW$"
	port := 20142

	db, err := xmysqlv2.NewDBBuilder(&xmysql.Database{
		Host: host,
		UserName: user,
		Password: password,
		Port: port,
		DBName: dbname,
	}, &xmysqlv2.ServiceInfo{
		LocalEnv:    "local",
		DeployEnv:   "prod",
		ServiceName: "logic",
	}).WithNameSuffix("master").Build(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	var databases []string
	err = db.Query(&databases, "SHOW DATABASES")
	if err != nil {
		t.Fatalf("Failed to get databases: %v", err)
	}
	t.Logf("Databases: %+v", databases)

	var tables []string
	err = db.Query(&tables, "SHOW TABLES")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}
	t.Logf("Tables: %+v", tables)

	var schema string
	err = db.Query(&schema, "SHOW CREATE TABLE `kim_partition_dbinst`")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}
	t.Logf("Schema: %+v", schema)
}

func TestOpenguassDialect_getColumnsFromSchema(t *testing.T) {
	// 创建一个 OpenguassDialect 实例用于测试（不需要真实的数据库连接）
	dialect := &OpenguassDialect{BaseDialect: &BaseDialect{}}

	tests := []struct {
		name          string
		schema        string
		expectedCols  []ColumnInfo
		expectedError bool
	}{
		{
			name: "测试用例：kim_event_chat 表",
			schema: `CREATE TABLE kim_event_chat (
	id bigint  NOT NULL DEFAULT nextval('test_kim_1164367684788410899.kim_event_chat_id_seq'::regclass),
	chatid bigint  NOT NULL DEFAULT 0,
	ver bigint  NOT NULL DEFAULT 0,
	seq bigint  NOT NULL DEFAULT 0,
	op_type bigint  NOT NULL DEFAULT 0,
	data bytea  NOT NULL DEFAULT '\x'::bytea,
	PRIMARY KEY  (id,id )
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('test_kim_1164367684788410899.kim_event_chat_id_seq'::regclass)", Key: "PRI"},
				{Name: "chatid", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "ver", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "seq", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "op_type", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "data", Type: "bytea", DefaultValue: "'\\x'", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试包含字符串 DEFAULT 值的列",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL DEFAULT nextval('test_seq'::regclass),
	name varchar(100) NOT NULL DEFAULT 'test',
	email varchar(255) NOT NULL DEFAULT '',
	status varchar(20) NOT NULL DEFAULT 'active',
	PRIMARY KEY (id)
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('test_seq'::regclass)", Key: "PRI"},
				{Name: "name", Type: "varchar(100)", DefaultValue: "'test'", Key: ""},
				{Name: "email", Type: "varchar(255)", DefaultValue: "''", Key: ""},
				{Name: "status", Type: "varchar(20)", DefaultValue: "'active'", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试数字 DEFAULT 值",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL DEFAULT nextval('seq'::regclass),
	age integer NOT NULL DEFAULT 18,
	score numeric(10,2) NOT NULL DEFAULT 0.00,
	count bigint NOT NULL DEFAULT 0,
	PRIMARY KEY (id)
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('seq'::regclass)", Key: "PRI"},
				{Name: "age", Type: "integer", DefaultValue: "18", Key: ""},
				{Name: "score", Type: "numeric(10,2)", DefaultValue: "0.00", Key: ""},
				{Name: "count", Type: "bigint", DefaultValue: "0", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试空表结构",
			schema: `CREATE TABLE empty_table (
);`,
			expectedCols: []ColumnInfo{},
			expectedError: false,
		},
		{
			name: "测试没有 DEFAULT 值的列",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL,
	name varchar(100) NOT NULL,
	created_at timestamp NOT NULL
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "", Key: ""},
				{Name: "name", Type: "varchar(100)", DefaultValue: "", Key: ""},
				{Name: "created_at", Type: "timestamp", DefaultValue: "", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试混合情况：有 DEFAULT 和没有 DEFAULT 的列",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL DEFAULT nextval('seq'::regclass),
	name varchar(100) NOT NULL,
	status varchar(20) NOT NULL DEFAULT 'active',
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL DEFAULT '2024-01-01',
	PRIMARY KEY (id)
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('seq'::regclass)", Key: "PRI"},
				{Name: "name", Type: "varchar(100)", DefaultValue: "", Key: ""},
				{Name: "status", Type: "varchar(20)", DefaultValue: "'active'", Key: ""},
				{Name: "created_at", Type: "timestamp", DefaultValue: "", Key: ""},
				{Name: "updated_at", Type: "timestamp", DefaultValue: "'2024-01-01'", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试复杂类型定义",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL DEFAULT nextval('seq'::regclass),
	data varchar(100) NOT NULL DEFAULT 'default',
	value numeric(10,2) NOT NULL DEFAULT 0.00,
	content text NOT NULL DEFAULT '',
	PRIMARY KEY (id)
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('seq'::regclass)", Key: "PRI"},
				{Name: "data", Type: "varchar(100)", DefaultValue: "'default'", Key: ""},
				{Name: "value", Type: "numeric(10,2)", DefaultValue: "0.00", Key: ""},
				{Name: "content", Type: "text", DefaultValue: "''", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试多列主键",
			schema: `CREATE TABLE test_table (
	id bigint NOT NULL DEFAULT nextval('seq'::regclass),
	user_id bigint NOT NULL DEFAULT 0,
	role_id bigint NOT NULL DEFAULT 0,
	PRIMARY KEY (id, user_id)
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('seq'::regclass)", Key: "PRI"},
				{Name: "user_id", Type: "bigint", DefaultValue: "0", Key: "PRI"},
				{Name: "role_id", Type: "bigint", DefaultValue: "0", Key: ""},
			},
			expectedError: false,
		},
		{
			name: "测试用例：kim_partition_dbinst 表（实际场景）",
			schema: `CREATE TABLE kim_partition_dbinst (
	id bigint  NOT NULL DEFAULT nextval('test_kim_1164367684788410899.kim_partition_dbinst_id_seq'::regclass),
	src_database character varying (64) NOT NULL DEFAULT ''::character varying,
	src_table character varying (64) NOT NULL DEFAULT ''::character varying,
	mgr_id bigint  NOT NULL DEFAULT 0,
	db_inst character varying (64) NOT NULL DEFAULT ''::character varying,
	dst_table character varying (64) NOT NULL DEFAULT ''::character varying,
	start_pos bigint  NOT NULL DEFAULT 0,
	end_pos bigint  NOT NULL DEFAULT 0,
	PRIMARY KEY  (id,id,id )
);`,
			expectedCols: []ColumnInfo{
				{Name: "id", Type: "bigint", DefaultValue: "nextval('test_kim_1164367684788410899.kim_partition_dbinst_id_seq'::regclass)", Key: "PRI"},
				{Name: "src_database", Type: "character varying (64)", DefaultValue: "''", Key: ""},
				{Name: "src_table", Type: "character varying (64)", DefaultValue: "''", Key: ""},
				{Name: "mgr_id", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "db_inst", Type: "character varying (64)", DefaultValue: "''", Key: ""},
				{Name: "dst_table", Type: "character varying (64)", DefaultValue: "''", Key: ""},
				{Name: "start_pos", Type: "bigint", DefaultValue: "0", Key: ""},
				{Name: "end_pos", Type: "bigint", DefaultValue: "0", Key: ""},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接调用私有方法 getColumnsFromSchema
			columns, err := dialect.getColumnsFromSchema(tt.schema)

			// 验证错误
			if tt.expectedError {
				if err == nil {
					t.Errorf("期望返回错误，但实际没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望返回错误，但实际返回: %v", err)
				return
			}

			// 验证列数量
			if len(columns) != len(tt.expectedCols) {
				t.Errorf("列数量不匹配: 期望 %d, 实际 %d", len(tt.expectedCols), len(columns))
				t.Logf("期望的列: %+v", tt.expectedCols)
				t.Logf("实际的列: %+v", columns)
				return
			}

			// 验证每列的信息
			for i, expectedCol := range tt.expectedCols {
				if i >= len(columns) {
					t.Errorf("列索引 %d 超出范围", i)
					continue
				}

				actualCol := columns[i]
				if actualCol.Name != expectedCol.Name {
					t.Errorf("列 %d 名称不匹配: 期望 %s, 实际 %s", i, expectedCol.Name, actualCol.Name)
				}
				if actualCol.Type != expectedCol.Type {
					t.Errorf("列 %d 类型不匹配: 期望 %s, 实际 %s", i, expectedCol.Type, actualCol.Type)
				}
				if actualCol.DefaultValue != expectedCol.DefaultValue {
					t.Errorf("列 %d 默认值不匹配: 期望 %s, 实际 %s", i, expectedCol.DefaultValue, actualCol.DefaultValue)
				}
				if actualCol.Key != expectedCol.Key {
					t.Errorf("列 %d 主键标识不匹配: 期望 %s, 实际 %s", i, expectedCol.Key, actualCol.Key)
				}
			}
		})
	}
}