package database

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type BaseMysqlBasedDB struct {
	*MySQL
	dialect MysqlBasedDialect

	dialectType string
	dbConfig    *DBConfig
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func (m *DBConfig) BuildDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", m.User, m.Password, m.Host, m.Port, m.Database)
}

func GetDBConfigFromDSN(dsn string) (*DBConfig, error) {
	parsedDSN, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	host := parsedDSN.Addr
	addrs := strings.Split(parsedDSN.Addr, ":")
	port := 3306
	if len(addrs) > 1 {
		port, err = strconv.Atoi(addrs[1])
		if err != nil {
			return nil, err
		}
		host = addrs[0]
	}
	user := parsedDSN.User
	pwd, err := url.PathUnescape(parsedDSN.Passwd)
	if err != nil {
		return nil, err
	}
	password := pwd
	database := parsedDSN.DBName
	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}, nil
}

func NewBaseMysqlBasedDB(dialectType string) *BaseMysqlBasedDB {
	return &BaseMysqlBasedDB{dialectType: dialectType, MySQL: NewMySQL()}
}

// Connect 建立MySQL连接
func (m *BaseMysqlBasedDB) Connect(dsn string) error {
	err := m.MySQL.Connect(dsn)
	if err != nil {
		return err
	}
	dbConfig, err := GetDBConfigFromDSN(dsn)
	if err != nil {
		return err
	}
	m.dbConfig = dbConfig
	m.dialect = GetMysqlBasedDialectByType(m.dialectType, m.MySQL.db)
	return nil
}

func (m *BaseMysqlBasedDB) ConnectWithConfig(dbConfig *DBConfig) error {
	err := m.MySQL.Connect(dbConfig.BuildDSN())
	if err != nil {
		return err
	}
	m.dbConfig = dbConfig
	m.dialect = GetMysqlBasedDialectByType(m.dialectType, m.MySQL.db)
	return nil
}

// Close 关闭连接
func (m *BaseMysqlBasedDB) Close() error {
	if m.db != nil {
		m.db.Close()
	}
	return nil
}

// GetTables 获取所有表名
func (m *BaseMysqlBasedDB) GetTables() ([]string, error) {
	return m.dialect.GetTables()
}

// GetTableSchema 获取表结构
func (m *BaseMysqlBasedDB) GetTableSchema(tableName string) (string, error) {
	return m.dialect.GetTableSchema(tableName)
}

// GetTableColumns 获取表的列信息
func (m *BaseMysqlBasedDB) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	return m.dialect.GetTableColumns(tableName)
}

// GetDatabases 获取所有数据库名称
func (m *BaseMysqlBasedDB) GetDatabases() ([]string, error) {
	return m.dialect.GetDatabases()
}

// SwitchDatabase 切换当前使用的数据库
func (m *BaseMysqlBasedDB) SwitchDatabase(databaseName string) error {
	m.dbConfig.Database = databaseName
	return m.ConnectWithConfig(m.dbConfig)
}
