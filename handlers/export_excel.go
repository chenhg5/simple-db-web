package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportTableDataToExcel 导出表数据为Excel
func (s *Server) ExportTableDataToExcel(w http.ResponseWriter, r *http.Request) {
	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingTableName)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 50
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// 获取数据
	data, _, err := session.db.GetTableData(tableName, page, pageSize)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableDataFailed, err)
		return
	}

	// 获取列信息
	columns, err := session.db.GetTableColumns(tableName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeGetTableColumnsFailed, err)
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			s.getLogger().Error(r.Context(), "Failed to close Excel file: %v", err)
		}
	}()

	sheetName := "Sheet1"
	// 删除默认的Sheet1（如果存在）
	val, _ := f.GetSheetIndex("Sheet1")
	if val != -1 {
		f.DeleteSheet("Sheet1")
	}
	index, err := f.NewSheet(sheetName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeCreateExcelSheetFailed, err)
		return
	}
	f.SetActiveSheet(index)

	// 写入表头
	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = col.Name
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, col.Name)
	}

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	if err == nil && len(columns) > 0 {
		lastCol, _ := excelize.CoordinatesToCellName(len(columns), 1)
		f.SetCellStyle(sheetName, "A1", lastCol, headerStyle)
	}

	// 写入数据
	for rowIdx, row := range data {
		for colIdx, colName := range colNames {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			value := row[colName]
			if value != nil {
				f.SetCellValue(sheetName, cellName, value)
			}
		}
	}

	// 设置响应头
	filename := fmt.Sprintf("%s_page%d_%s.xlsx", tableName, page, time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入响应
	if err := f.Write(w); err != nil {
		s.getLogger().Error(r.Context(), "Failed to write Excel file: %v", err)
		writeJSONError(w, http.StatusInternalServerError, ErrCodeExportExcelFailed, err)
		return
	}
}

// ExportQueryResultsToExcel 导出查询结果为Excel
func (s *Server) ExportQueryResultsToExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed)
		return
	}

	connectionID := getConnectionID(r)
	if connectionID == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeMissingConnectionID)
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeParseRequestFailed, err)
		return
	}

	if req.Query == "" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeEmptySQLQuery)
		return
	}

	session, err := s.getSession(connectionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, ErrCodeConnectionNotExists, err)
		return
	}

	// 只支持SELECT查询
	queryUpper := fmt.Sprintf("%.6s", req.Query)
	if queryUpper != "SELECT" && queryUpper != "select" {
		writeJSONError(w, http.StatusBadRequest, ErrCodeOnlySelectQueryAllowed)
		return
	}

	// 执行查询
	results, err := session.db.ExecuteQuery(req.Query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeExecuteQueryFailed, err)
		return
	}

	if len(results) == 0 {
		writeJSONError(w, http.StatusBadRequest, ErrCodeQueryResultEmpty)
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			s.getLogger().Error(r.Context(), "Failed to close Excel file: %v", err)
		}
	}()

	sheetName := "Sheet1"
	// 删除默认的Sheet1（如果存在）
	val, _ := f.GetSheetIndex("Sheet1")
	if val != -1 {
		f.DeleteSheet("Sheet1")
	}
	index, err := f.NewSheet(sheetName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, ErrCodeCreateExcelSheetFailed, err)
		return
	}
	f.SetActiveSheet(index)

	// 获取列名（从第一行数据中提取）
	colNames := make([]string, 0)
	if len(results) > 0 {
		for colName := range results[0] {
			colNames = append(colNames, colName)
		}
	}

	// 写入表头
	for i, colName := range colNames {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, colName)
	}

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	if err == nil && len(colNames) > 0 {
		lastCol, _ := excelize.CoordinatesToCellName(len(colNames), 1)
		f.SetCellStyle(sheetName, "A1", lastCol, headerStyle)
	}

	// 写入数据
	for rowIdx, row := range results {
		for colIdx, colName := range colNames {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			value := row[colName]
			if value != nil {
				f.SetCellValue(sheetName, cellName, value)
			}
		}
	}

	// 设置响应头
	filename := fmt.Sprintf("query_result_%s.xlsx", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入响应
	if err := f.Write(w); err != nil {
		s.getLogger().Error(r.Context(), "Failed to write Excel file: %v", err)
		writeJSONError(w, http.StatusInternalServerError, ErrCodeExportExcelFailed, err)
		return
	}
}
