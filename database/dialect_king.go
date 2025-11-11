package database

import "database/sql"

type KingbaseDialect struct {
	*BaseDialect
}

func NewKingbaseDialect(db *sql.DB) *KingbaseDialect {
	return &KingbaseDialect{BaseDialect: NewBaseDialect(db)}
}

func (m *KingbaseDialect) GetTableColumns(tableName string) ([]ColumnInfo, error) {
	schema, err := m.BaseDialect.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}
	return getColumnsFroPGLikeSchema(schema)
}
