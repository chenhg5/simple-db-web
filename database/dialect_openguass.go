package database

import (
	"regexp"
	"strings"

	"ksogit.kingsoft.net/kgo/mysql"
)

type OpenguassDialect struct {
	*BaseDialect
}

func NewOpenguassDialect(db mysql.DBAdapter) *OpenguassDialect {
	return &OpenguassDialect{BaseDialect: NewBaseDialect(db)}
}

func (m *OpenguassDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	schema, err := m.BaseDialect.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	return m.getColumnsFromSchema(schema)
}

func (m *OpenguassDialect) getColumnsFromSchema(schema string) ([]ColumnInfo, error) {
	columns := []ColumnInfo{}

	// 首先解析主键列名
	primaryKeys := make(map[string]bool)
	// 匹配 PRIMARY KEY (col1, col2) 或 PRIMARY KEY  (col1)
	pkReg := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
	pkMatches := pkReg.FindAllStringSubmatch(schema, -1)
	for _, pkMatch := range pkMatches {
		// 解析主键列名列表，可能是 id 或 id,id 等（去除空格）
		colNamesStr := strings.TrimSpace(pkMatch[1])
		colNames := strings.FieldsFunc(colNamesStr, func(r rune) bool {
			return r == ',' || r == ' '
		})
		for _, colName := range colNames {
			colName = strings.TrimSpace(colName)
			if colName != "" {
				primaryKeys[colName] = true
			}
		}
	}

	// 匹配列定义
	// 格式: column_name type NOT NULL DEFAULT value
	// 或: column_name type NOT NULL
	// 注意：列名没有引号，类型可能包含空格（如 character varying (64)）
	// DEFAULT 值可能是函数调用、字符串、数字等
	// 先匹配列名和类型，然后手动解析 DEFAULT 值（因为可能包含括号）
	reg := regexp.MustCompile(`(\w+)\s+(.+?)\s+NOT\s+NULL`)
	matches := reg.FindAllStringSubmatch(schema, -1)

	for _, match := range matches {
		col := ColumnInfo{
			Name: match[1],
			Type: strings.TrimSpace(match[2]),
		}

		// 判断是否是主键
		if primaryKeys[col.Name] {
			col.Key = "PRI"
		}

		// 查找该列定义的完整文本，用于提取 DEFAULT 值
		// 找到列定义在 schema 中的位置（使用更精确的匹配）
		colPattern := regexp.MustCompile(regexp.QuoteMeta(match[0]))
		colDefMatches := colPattern.FindAllStringIndex(schema, -1)
		var colDefStart int = -1
		for _, idx := range colDefMatches {
			// 检查这个位置是否真的是列定义的开始（前面应该是换行、空格或逗号）
			before := ""
			if idx[0] > 0 {
				before = schema[idx[0]-1:idx[0]]
			}
			if idx[0] == 0 || before == "\n" || before == "\t" || before == " " || before == "," || before == "(" {
				colDefStart = idx[0]
				break
			}
		}
		
		if colDefStart >= 0 {
			// 从列定义开始查找 DEFAULT（在当前列定义范围内）
			remaining := schema[colDefStart+len(match[0]):]
			// 找到下一个列定义或 PRIMARY KEY 或表结束的位置
			nextColPattern := regexp.MustCompile(`(\w+)\s+\S+\s+NOT\s+NULL|PRIMARY\s+KEY|\)\s*;`)
			nextColMatch := nextColPattern.FindStringIndex(remaining)
			colDefEnd := len(remaining)
			if nextColMatch != nil {
				colDefEnd = nextColMatch[0]
			}
			colDefText := remaining[:colDefEnd]
			
			defaultIdx := strings.Index(colDefText, "DEFAULT")
			if defaultIdx >= 0 {
				// 找到 DEFAULT 关键字，提取默认值
				defaultStart := defaultIdx + len("DEFAULT")
				defaultValueStr := strings.TrimSpace(colDefText[defaultStart:])
				
				// 手动解析默认值，正确处理括号嵌套
				var defaultValue strings.Builder
				parenCount := 0
				inString := false
				stringChar := byte(0)
				
				for i := 0; i < len(defaultValueStr); i++ {
					char := defaultValueStr[i]
					
					// 处理字符串
					if !inString && (char == '\'' || char == '"') {
						inString = true
						stringChar = char
						defaultValue.WriteByte(char)
					} else if inString && char == stringChar {
						// 检查是否是转义的引号
						if i > 0 && defaultValueStr[i-1] == '\\' {
							defaultValue.WriteByte(char)
						} else {
							inString = false
							defaultValue.WriteByte(char)
						}
					} else {
						defaultValue.WriteByte(char)
						
						// 处理括号
						if !inString {
							if char == '(' {
								parenCount++
							} else if char == ')' {
								parenCount--
								// 如果括号闭合，且下一个字符是逗号或右括号，则结束
								if parenCount == 0 {
									// 检查后面是否是逗号、右括号或类型转换
									nextPart := strings.TrimSpace(defaultValueStr[i+1:])
									if strings.HasPrefix(nextPart, ",") || strings.HasPrefix(nextPart, ")") {
										break
									}
									// 如果是类型转换 ::type，继续读取
									if strings.HasPrefix(nextPart, "::") {
										// 找到类型转换的结束位置（逗号或右括号）
										typeEnd := strings.IndexAny(nextPart, ",)")
										if typeEnd > 0 {
											defaultValue.WriteString(nextPart[:typeEnd])
											break
										}
									}
								}
							} else if parenCount == 0 && (char == ',' || char == ')') {
								// 没有括号，遇到逗号或右括号就结束
								// 回退一个字符（不包括逗号或右括号）
								defaultValue.Reset()
								defaultValue.WriteString(strings.TrimRight(defaultValueStr[:i], " ,"))
								break
							}
						}
					}
				}
				
				if defaultValue.Len() > 0 {
					defaultValueStr := strings.TrimSpace(defaultValue.String())
					// 移除末尾的逗号（如果有）
					defaultValueStr = strings.TrimRight(defaultValueStr, ",")
					
					// 处理类型转换，如 '\x'::bytea -> '\x'
					// 但保留函数调用中的类型转换，如 nextval('...'::regclass) 应该保留
					if strings.Contains(defaultValueStr, "::") {
						// 检查是否是函数调用（包含括号）
						if strings.Contains(defaultValueStr, "(") && strings.Contains(defaultValueStr, ")") {
							// 函数调用，保留完整的表达式
							col.DefaultValue = defaultValueStr
						} else {
							// 简单的类型转换，移除类型部分
							parts := strings.Split(defaultValueStr, "::")
							defaultValueStr = strings.TrimSpace(parts[0])
							col.DefaultValue = defaultValueStr
						}
					} else {
						col.DefaultValue = defaultValueStr
					}
				}
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}