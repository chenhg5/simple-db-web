package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// RequireLimitValidator 要求SELECT查询必须包含LIMIT的校验器
type RequireLimitValidator struct{}

// NewRequireLimitValidator 创建RequireLimitValidator实例
func NewRequireLimitValidator() *RequireLimitValidator {
	return &RequireLimitValidator{}
}

// Name 返回校验器名称
func (v *RequireLimitValidator) Name() string {
	return "RequireLimit"
}

// Validate 校验SELECT查询是否包含LIMIT
func (v *RequireLimitValidator) Validate(query string, queryType string) error {
	// 只对SELECT查询进行校验
	if queryType != "SELECT" {
		return nil
	}

	// 转换为大写进行匹配
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	
	// 检查是否包含LIMIT关键字（使用单词边界确保匹配完整单词）
	limitPattern := regexp.MustCompile(`\bLIMIT\s+`)
	if !limitPattern.MatchString(queryUpper) {
		return fmt.Errorf("SELECT查询必须包含LIMIT子句以限制返回行数")
	}

	return nil
}

// NoDropTableValidator 禁止DROP TABLE语句的校验器
type NoDropTableValidator struct{}

// NewNoDropTableValidator 创建NoDropTableValidator实例
func NewNoDropTableValidator() *NoDropTableValidator {
	return &NoDropTableValidator{}
}

// Name 返回校验器名称
func (v *NoDropTableValidator) Name() string {
	return "NoDropTable"
}

// Validate 禁止DROP TABLE语句
func (v *NoDropTableValidator) Validate(query string, queryType string) error {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	
	// 检查是否包含DROP TABLE
	dropTablePattern := regexp.MustCompile(`\bDROP\s+TABLE\b`)
	if dropTablePattern.MatchString(queryUpper) {
		return fmt.Errorf("不允许执行DROP TABLE语句")
	}

	return nil
}

// NoTruncateValidator 禁止TRUNCATE语句的校验器
type NoTruncateValidator struct{}

// NewNoTruncateValidator 创建NoTruncateValidator实例
func NewNoTruncateValidator() *NoTruncateValidator {
	return &NoTruncateValidator{}
}

// Name 返回校验器名称
func (v *NoTruncateValidator) Name() string {
	return "NoTruncate"
}

// Validate 禁止TRUNCATE语句
func (v *NoTruncateValidator) Validate(query string, queryType string) error {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	
	// 检查是否包含TRUNCATE TABLE
	truncatePattern := regexp.MustCompile(`\bTRUNCATE\s+TABLE\b`)
	if truncatePattern.MatchString(queryUpper) {
		return fmt.Errorf("不允许执行TRUNCATE TABLE语句")
	}
	
	// 也检查TRUNCATE（不带TABLE）
	truncateOnlyPattern := regexp.MustCompile(`^\s*TRUNCATE\s+`)
	if truncateOnlyPattern.MatchString(queryUpper) {
		return fmt.Errorf("不允许执行TRUNCATE语句")
	}

	return nil
}

// NoDropDatabaseValidator 禁止DROP DATABASE语句的校验器
type NoDropDatabaseValidator struct{}

// NewNoDropDatabaseValidator 创建NoDropDatabaseValidator实例
func NewNoDropDatabaseValidator() *NoDropDatabaseValidator {
	return &NoDropDatabaseValidator{}
}

// Name 返回校验器名称
func (v *NoDropDatabaseValidator) Name() string {
	return "NoDropDatabase"
}

// Validate 禁止DROP DATABASE语句
func (v *NoDropDatabaseValidator) Validate(query string, queryType string) error {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	
	// 检查是否包含DROP DATABASE
	dropDbPattern := regexp.MustCompile(`\bDROP\s+DATABASE\b`)
	if dropDbPattern.MatchString(queryUpper) {
		return fmt.Errorf("不允许执行DROP DATABASE语句")
	}

	return nil
}

// MaxQueryLengthValidator 限制查询最大长度的校验器
type MaxQueryLengthValidator struct {
	MaxLength int
}

// NewMaxQueryLengthValidator 创建MaxQueryLengthValidator实例
// maxLength: 允许的最大查询长度（字符数）
func NewMaxQueryLengthValidator(maxLength int) *MaxQueryLengthValidator {
	return &MaxQueryLengthValidator{
		MaxLength: maxLength,
	}
}

// Name 返回校验器名称
func (v *MaxQueryLengthValidator) Name() string {
	return "MaxQueryLength"
}

// Validate 校验查询长度
func (v *MaxQueryLengthValidator) Validate(query string, queryType string) error {
	if len(query) > v.MaxLength {
		return fmt.Errorf("查询长度超过限制（最大%d字符）", v.MaxLength)
	}
	return nil
}

