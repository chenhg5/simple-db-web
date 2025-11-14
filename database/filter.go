package database

import (
	"fmt"
	"strings"
)

// BuildWhereClause 构建 WHERE 子句
// dbType: 数据库类型（mysql, postgresql, sqlite, sqlserver, oracle等）
// tableName: 表名（用于标识符引用）
// filters: 过滤条件组
// 返回: WHERE 子句字符串（不包含 WHERE 关键字）和参数列表
func BuildWhereClause(dbType string, tableName string, filters *FilterGroup) (string, []interface{}, error) {
	if filters == nil || len(filters.Conditions) == 0 {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}

	// 确定逻辑关系（默认为 AND）
	logic := strings.ToUpper(filters.Logic)
	if logic != "AND" && logic != "OR" {
		logic = "AND"
	}

	// 获取标识符引用方式和占位符函数
	quoteFunc := getQuoteFunc(dbType)
	placeholderFunc := getPlaceholderFunc(dbType)

	argIndex := 1 // PostgreSQL 从 $1 开始

	for _, condition := range filters.Conditions {
		if condition.Field == "" {
			continue
		}

		// 引用字段名
		quotedField := quoteFunc(condition.Field)

		operator := strings.ToUpper(strings.TrimSpace(condition.Operator))
		
		switch operator {
		case "=", "!=", "<", ">", "<=", ">=":
			// 简单的比较操作符
			if condition.Value != "" {
				placeholder := placeholderFunc(argIndex)
				conditions = append(conditions, fmt.Sprintf("%s %s %s", quotedField, operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		
		case "LIKE", "NOT LIKE":
			// LIKE 操作符
			if condition.Value != "" {
				placeholder := placeholderFunc(argIndex)
				conditions = append(conditions, fmt.Sprintf("%s %s %s", quotedField, operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		
		case "IN", "NOT IN":
			// IN 操作符
			values := condition.Values
			if len(values) == 0 && condition.Value != "" {
				// 如果提供了 Value 字符串，按逗号分割
				values = strings.Split(condition.Value, ",")
				// 去除每个值的空白
				for i := range values {
					values[i] = strings.TrimSpace(values[i])
				}
			}
			
			if len(values) > 0 {
				placeholders := make([]string, len(values))
				for i := range placeholders {
					placeholders[i] = placeholderFunc(argIndex)
					args = append(args, values[i])
					argIndex++
				}
				conditions = append(conditions, fmt.Sprintf("%s %s (%s)", quotedField, operator, strings.Join(placeholders, ", ")))
			}
		
		case "IS NULL", "IS NOT NULL":
			// NULL 检查（不需要参数）
			conditions = append(conditions, fmt.Sprintf("%s %s", quotedField, operator))
		
		default:
			// 默认使用 = 操作符
			if condition.Value != "" {
				placeholder := placeholderFunc(argIndex)
				conditions = append(conditions, fmt.Sprintf("%s = %s", quotedField, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		}
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	whereClause := strings.Join(conditions, " "+logic+" ")
	return whereClause, args, nil
}

// getPlaceholderFunc 根据数据库类型返回占位符函数
func getPlaceholderFunc(dbType string) func(int) string {
	switch dbType {
	case "postgresql":
		return func(index int) string {
			return fmt.Sprintf("$%d", index)
		}
	default:
		// MySQL, SQLite, SQL Server, Oracle 等都使用 ?
		return func(index int) string {
			return "?"
		}
	}
}

// getQuoteFunc 根据数据库类型返回字段引用函数
func getQuoteFunc(dbType string) func(string) string {
	switch dbType {
	case "mysql", "sqlite", "h2":
		return func(s string) string {
			return fmt.Sprintf("`%s`", s)
		}
	case "postgresql":
		return func(s string) string {
			return fmt.Sprintf(`"%s"`, s)
		}
	case "sqlserver":
		return func(s string) string {
			return fmt.Sprintf("[%s]", s)
		}
	case "oracle":
		return func(s string) string {
			return fmt.Sprintf(`"%s"`, strings.ToUpper(s))
		}
	case "clickhouse":
		return func(s string) string {
			return fmt.Sprintf("`%s`", s)
		}
	default:
		// 默认使用反引号
		return func(s string) string {
			return fmt.Sprintf("`%s`", s)
		}
	}
}

