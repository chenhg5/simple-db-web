package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type OceandbDialect struct {
	*BaseDialect
}

func NewOceandbDialect(db *sql.DB) *OceandbDialect {
	return &OceandbDialect{BaseDialect: NewBaseDialect(db)}
}

func (m *OceandbDialect) GetTableSchema(tableName string) (string, error) {
	type CreateTable struct {
		Table       string `db:"Table"`
		CreateTable string `db:"Create Table"`
	}
	var createTables = make([]CreateTable, 0)
	rows, err := m.db.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
	if err != nil {
		return "", fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var createTable CreateTable
		if err := rows.Scan(&createTable.Table, &createTable.CreateTable); err != nil {
			return "", err
		}
		createTables = append(createTables, createTable)
	}

	return createTables[0].CreateTable, nil
}

func (m *OceandbDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	schema, err := m.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	return getColumnsFroMySQLLikeSchema(schema)
}

func getColumnsFroMySQLLikeSchema(schema string) ([]ColumnInfo, error) {
	columns := []ColumnInfo{}

	// 首先解析主键列名
	primaryKeys := make(map[string]bool)
	// 匹配 PRIMARY KEY (`col1`, `col2`) 或 PRIMARY KEY (`col1`)
	pkReg := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
	pkMatches := pkReg.FindAllStringSubmatch(schema, -1)
	for _, pkMatch := range pkMatches {
		// 解析主键列名列表，列名可能用反引号包裹
		colNamesStr := strings.TrimSpace(pkMatch[1])
		// 匹配反引号内的列名
		colNameReg := regexp.MustCompile("`([^`]+)`")
		colNameMatches := colNameReg.FindAllStringSubmatch(colNamesStr, -1)
		for _, colNameMatch := range colNameMatches {
			if len(colNameMatch) > 1 {
				primaryKeys[colNameMatch[1]] = true
			}
		}
	}

	// 匹配列定义
	// 格式: `column_name` type [unsigned] [NOT NULL] [DEFAULT 'value'] [AUTO_INCREMENT] [COMMENT 'comment']
	// 注意：列名用反引号包裹，类型可能包含括号和修饰符（如 bigint(20) unsigned）
	// 先找到 CREATE TABLE 后的括号内容
	tableStart := strings.Index(schema, "CREATE TABLE")
	if tableStart < 0 {
		return columns, nil
	}

	// 找到表定义的开始括号
	parenStart := strings.Index(schema[tableStart:], "(")
	if parenStart < 0 {
		return columns, nil
	}
	tableDefStart := tableStart + parenStart + 1

	// 找到表定义的结束括号
	parenCount := 1
	tableDefEnd := tableDefStart
	for i := tableDefStart; i < len(schema); i++ {
		if schema[i] == '(' {
			parenCount++
		} else if schema[i] == ')' {
			parenCount--
			if parenCount == 0 {
				tableDefEnd = i
				break
			}
		}
	}

	// 只处理表定义括号内的内容
	tableDef := schema[tableDefStart:tableDefEnd]

	// 匹配列定义（在表定义括号内）
	colDefPattern := regexp.MustCompile("`([^`]+)`\\s+([^`]+?)(?:\\s+NOT\\s+NULL|\\s+AUTO_INCREMENT|\\s+DEFAULT|\\s+COMMENT|,|\\n|$)")
	allMatches := colDefPattern.FindAllStringSubmatch(tableDef, -1)

	for _, match := range allMatches {
		if len(match) < 3 {
			continue
		}

		colName := match[1]

		// 找到该列定义在 tableDef 中的位置
		colDefStart := strings.Index(tableDef, "`"+colName+"`")
		if colDefStart < 0 {
			continue
		}

		// 检查前面是否有 UNIQUE KEY 或 KEY，如果有则跳过（这是索引定义，不是列定义）
		beforeText := tableDef[:colDefStart]
		if strings.Contains(strings.ToLower(beforeText), "unique key") ||
			(strings.Contains(strings.ToLower(beforeText), "key `") && !strings.Contains(strings.ToLower(beforeText), "primary key")) {
			continue
		}

		// 检查是否是表名（CREATE TABLE `table_name` 中的表名）
		// 如果前面是 CREATE TABLE，则跳过
		beforeLower := strings.ToLower(beforeText)
		if strings.HasSuffix(beforeLower, "create table") || strings.HasSuffix(beforeLower, "create table ") {
			continue
		}

		col := ColumnInfo{
			Name: colName,
		}

		// 查找该列定义的完整文本，用于提取类型、DEFAULT 值、AUTO_INCREMENT 等
		// 找到下一个列定义或 PRIMARY KEY、UNIQUE KEY、KEY 或表结束的位置
		remaining := tableDef[colDefStart:]
		nextColPattern := regexp.MustCompile(`,\s*` + regexp.QuoteMeta("`") + `|PRIMARY\s+KEY|UNIQUE\s+KEY|KEY\s+` + regexp.QuoteMeta("`") + `|\)\s*;|\)\s*$`)
		nextColMatch := nextColPattern.FindStringIndex(remaining)
		colDefEnd := len(remaining)
		if nextColMatch != nil {
			colDefEnd = nextColMatch[0]
		}
		colDefText := strings.TrimSpace(remaining[:colDefEnd])

		// 提取列定义部分（去掉列名）
		colDefOnly := strings.TrimSpace(strings.TrimPrefix(colDefText, "`"+colName+"`"))

		// 提取类型（到 NOT NULL、DEFAULT、AUTO_INCREMENT、COMMENT 之前）
		// 类型可能包含括号，如 bigint(20) unsigned
		typeEndPattern := regexp.MustCompile(`\s+(?:NOT\s+NULL|DEFAULT|AUTO_INCREMENT|COMMENT)`)
		typeEndMatch := typeEndPattern.FindStringIndex(colDefOnly)
		if typeEndMatch != nil {
			col.Type = strings.TrimSpace(colDefOnly[:typeEndMatch[0]])
		} else {
			col.Type = strings.TrimSpace(colDefOnly)
		}

		// 检查是否有 AUTO_INCREMENT（通常表示主键）
		if strings.Contains(colDefOnly, "AUTO_INCREMENT") {
			col.Key = "PRI"
		}

		// 判断是否是主键（从 PRIMARY KEY 定义中）
		if primaryKeys[col.Name] {
			col.Key = "PRI"
		}

		// 提取 DEFAULT 值
		// 匹配 DEFAULT 'value' 或 DEFAULT value
		defaultReg := regexp.MustCompile(`DEFAULT\s+('([^']*)'|([^\s,]+))`)
		defaultMatch := defaultReg.FindStringSubmatch(colDefOnly)
		if defaultMatch != nil {
			if defaultMatch[2] != "" {
				// 单引号内的值
				col.DefaultValue = defaultMatch[2]
			} else if defaultMatch[3] != "" {
				// 数字或其他值（如 CURRENT_TIMESTAMP）
				col.DefaultValue = defaultMatch[3]
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}
