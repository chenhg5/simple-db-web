package database

import (
	"database/sql"
	"regexp"
	"strings"
)

type DamengDialect struct {
	*BaseDialect
}

func NewDamengDialect(db *sql.DB) *DamengDialect {
	return &DamengDialect{BaseDialect: NewBaseDialect(db)}
}

func (m *DamengDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	schema, err := m.BaseDialect.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	return m.getColumnsFromSchema(schema)
}

func (m *DamengDialect) getColumnsFromSchema(schema string) ([]ColumnInfo, error) {
	columns := []ColumnInfo{}
	
	// 匹配 IDENTITY 模式，用于识别主键
	identityReg := regexp.MustCompile(`IDENTITY\s*\([^)]+\)`)
	
	// 匹配所有列定义（包括有 DEFAULT 和没有 DEFAULT 的）
	// 格式1: "column_name" TYPE IDENTITY(1, 1) NOT NULL,
	// 格式2: "column_name" TYPE DEFAULT 'value' NOT NULL,
	// 格式3: "column_name" TYPE DEFAULT 0 NOT NULL,
	// 注意：使用 (.+?) 非贪婪匹配，匹配到 DEFAULT 或 NOT NULL 之前
	reg := regexp.MustCompile(`"([^"]+)"\s+(.+?)(?:\s+DEFAULT\s+(?:'([^']*)'|(\S+)))?\s+NOT\s+NULL`)
	matches := reg.FindAllStringSubmatch(schema, -1)
	
	for _, match := range matches {
		col := ColumnInfo{
			Name: match[1],
		}
		
		typeAndAttrs := strings.TrimSpace(match[2])
		
		// 检查是否包含 IDENTITY，如果包含则标记为主键
		if identityReg.MatchString(typeAndAttrs) {
			col.Key = "PRI"
			// 从类型字符串中移除 IDENTITY 部分
			typeAndAttrs = identityReg.ReplaceAllString(typeAndAttrs, "")
			typeAndAttrs = strings.TrimSpace(typeAndAttrs)
			// 移除 IDENTITY 后的多余空格
			typeAndAttrs = strings.Trim(typeAndAttrs, " ")
		}
		col.Type = typeAndAttrs
		
		// 默认值可能是单引号内的值（match[3]）或数字值（match[4]）
		if match[3] != "" {
			col.DefaultValue = match[3]
		} else if match[4] != "" {
			col.DefaultValue = match[4]
		}
		
		columns = append(columns, col)
	}
	
	return columns, nil
}