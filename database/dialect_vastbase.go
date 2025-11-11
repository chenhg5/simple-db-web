package database

import "database/sql"

type VastbaseDialect struct {
	*BaseDialect
}

func NewVastbaseDialect(db *sql.DB) *VastbaseDialect {
	return &VastbaseDialect{BaseDialect: NewBaseDialect(db)}
}

func (m *VastbaseDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	schema, err := m.BaseDialect.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	return getColumnsFroPGLikeSchema(schema)
}
