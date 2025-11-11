// 全局状态
let currentTable = null;
let currentPage = 1;
let pageSize = 50;
let currentRowData = null;
let currentDeleteWhere = null;

// DOM元素
const connectionStatus = document.getElementById('connectionStatus');
const connectionForm = document.getElementById('connectionForm');
const connectionMode = document.getElementById('connectionMode');
const dsnGroup = document.getElementById('dsnGroup');
const formGroup = document.getElementById('formGroup');
const connectionPanel = document.getElementById('connectionPanel');
const databasePanel = document.getElementById('databasePanel');
const databaseSelect = document.getElementById('databaseSelect');
const disconnectBtn = document.getElementById('disconnectBtn');
const tablesPanel = document.getElementById('tablesPanel');
const tableList = document.getElementById('tableList');
const refreshTables = document.getElementById('refreshTables');
const tableFilter = document.getElementById('tableFilter');
const tabs = document.querySelectorAll('.tab');
const tabContents = document.querySelectorAll('.tab-content');
const dataTab = document.getElementById('dataTab');
const schemaTab = document.getElementById('schemaTab');
const queryTab = document.getElementById('queryTab');
const dataTableHead = document.getElementById('dataTableHead');
const dataTableBody = document.getElementById('dataTableBody');
const refreshData = document.getElementById('refreshData');
const pagination = document.getElementById('pagination');
const paginationInfo = document.getElementById('paginationInfo');
const schemaContent = document.getElementById('schemaContent');
const sqlQuery = document.getElementById('sqlQuery');
const executeQuery = document.getElementById('executeQuery');
const clearQuery = document.getElementById('clearQuery');
const queryResults = document.getElementById('queryResults');
const editModal = document.getElementById('editModal');
const editForm = document.getElementById('editForm');
const closeEditModal = document.getElementById('closeEditModal');
const cancelEdit = document.getElementById('cancelEdit');
const saveEdit = document.getElementById('saveEdit');
const deleteModal = document.getElementById('deleteModal');
const closeDeleteModal = document.getElementById('closeDeleteModal');
const cancelDelete = document.getElementById('cancelDelete');
const confirmDelete = document.getElementById('confirmDelete');

// 连接模式切换
connectionMode.addEventListener('change', (e) => {
    if (e.target.value === 'dsn') {
        dsnGroup.style.display = 'block';
        formGroup.style.display = 'none';
    } else {
        dsnGroup.style.display = 'none';
        formGroup.style.display = 'block';
    }
});

// 连接数据库
connectionForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const mode = connectionMode.value;
    const dbType = document.getElementById('dbType').value;
    
    let connectionInfo = {
        type: dbType
    };
    
    if (mode === 'dsn') {
        connectionInfo.dsn = document.getElementById('dsn').value;
    } else {
        connectionInfo.host = document.getElementById('host').value;
        connectionInfo.port = document.getElementById('port').value || '3306';
        connectionInfo.user = document.getElementById('user').value;
        connectionInfo.password = document.getElementById('password').value;
        // 不指定数据库,连接后让用户选择
        connectionInfo.database = '';
    }
    
    try {
        const response = await fetch('/api/connect', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(connectionInfo)
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            updateConnectionStatus(true);
            // 检查DSN中是否包含数据库
            const dsn = mode === 'dsn' ? document.getElementById('dsn').value : '';
            const hasDatabaseInDSN = dsn && (dsn.includes('/') && !dsn.endsWith('/') && !dsn.endsWith('/?'));
            
            if (hasDatabaseInDSN) {
                // DSN中包含数据库,直接使用该数据库
                connectionPanel.style.display = 'none';
                databasePanel.style.display = 'block';
                await loadDatabases(data.databases || []);
                // 尝试从DSN中提取数据库名
                const dbMatch = dsn.match(/\/([^\/\?]+)/);
                if (dbMatch && dbMatch[1]) {
                    const dbName = dbMatch[1];
                    // 设置选择器并切换数据库
                    databaseSelect.value = dbName;
                    await switchDatabase(dbName);
                } else {
                    await loadTables();
                }
            } else {
                // DSN中不包含数据库,显示数据库选择器
                connectionPanel.style.display = 'none';
                databasePanel.style.display = 'block';
                await loadDatabases(data.databases || []);
            }
            showNotification('连接成功', 'success');
        } else {
            showNotification(data.message || '连接失败', 'error');
        }
    } catch (error) {
        showNotification('连接失败: ' + error.message, 'error');
    }
});

// 更新连接状态
function updateConnectionStatus(connected) {
    const indicator = connectionStatus.querySelector('.status-indicator');
    const text = connectionStatus.querySelector('span:last-child');
    
    if (connected) {
        indicator.classList.add('connected');
        indicator.classList.remove('disconnected');
        text.textContent = '已连接';
    } else {
        indicator.classList.remove('connected');
        indicator.classList.add('disconnected');
        text.textContent = '未连接';
    }
}

// 加载数据库列表
async function loadDatabases(databases) {
    databaseSelect.innerHTML = '<option value="">请选择数据库...</option>';
    if (databases && databases.length > 0) {
        databases.forEach(db => {
            const option = document.createElement('option');
            option.value = db;
            option.textContent = db;
            databaseSelect.appendChild(option);
        });
    } else {
        // 如果没有数据库列表,尝试从服务器获取
        try {
            const response = await fetch('/api/databases');
            const data = await response.json();
            if (data.success && data.databases) {
                data.databases.forEach(db => {
                    const option = document.createElement('option');
                    option.value = db;
                    option.textContent = db;
                    databaseSelect.appendChild(option);
                });
            }
        } catch (error) {
            showNotification('获取数据库列表失败: ' + error.message, 'error');
        }
    }
}

// 切换数据库函数
async function switchDatabase(databaseName) {
    if (!databaseName) {
        tablesPanel.style.display = 'none';
        currentTable = null;
        return;
    }
    
    try {
        const response = await fetch('/api/database/switch', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ database: databaseName })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showNotification('切换数据库成功', 'success');
            // 加载表列表
            if (data.tables) {
                displayTables(data.tables);
            } else {
                await loadTables();
            }
        } else {
            showNotification(data.message || '切换数据库失败', 'error');
        }
    } catch (error) {
        showNotification('切换数据库失败: ' + error.message, 'error');
    }
}

// 切换数据库
databaseSelect.addEventListener('change', async (e) => {
    await switchDatabase(e.target.value);
});

// 存储所有表名（用于筛选）
let allTables = [];

// 显示表列表
function displayTables(tables) {
    allTables = tables;
    filterTables();
    tablesPanel.style.display = 'block';
}

// 筛选表列表
function filterTables() {
    const filterText = tableFilter.value.trim();
    const filteredTables = filterText 
        ? allTables.filter(table => table.toLowerCase().startsWith(filterText.toLowerCase()))
        : allTables;
    
    tableList.innerHTML = '';
    if (filteredTables.length === 0) {
        tableList.innerHTML = '<li style="padding: 1rem; color: var(--text-secondary);">没有找到表</li>';
    } else {
        filteredTables.forEach(table => {
            const li = document.createElement('li');
            li.className = 'table-item';
            li.textContent = table;
            li.addEventListener('click', () => selectTable(table));
            tableList.appendChild(li);
        });
    }
}

// 表筛选输入框事件
tableFilter.addEventListener('input', filterTables);

// 断开连接
disconnectBtn.addEventListener('click', async () => {
    try {
        const response = await fetch('/api/disconnect', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            updateConnectionStatus(false);
            // 显示连接表单,隐藏数据库选择器
            connectionPanel.style.display = 'block';
            databasePanel.style.display = 'none';
            tablesPanel.style.display = 'none';
            currentTable = null;
            databaseSelect.innerHTML = '<option value="">请选择数据库...</option>';
            // 清空筛选框和表列表
            tableFilter.value = '';
            allTables = [];
            currentColumns = [];
            showNotification('已断开连接', 'success');
        } else {
            showNotification(data.message || '断开连接失败', 'error');
        }
    } catch (error) {
        showNotification('断开连接失败: ' + error.message, 'error');
    }
});

// 加载表列表
async function loadTables() {
    try {
        const response = await fetch('/api/tables');
        const data = await response.json();
        
        if (data.success) {
            displayTables(data.tables || []);
        }
    } catch (error) {
        showNotification('加载表列表失败: ' + error.message, 'error');
    }
}

// 刷新表列表
refreshTables.addEventListener('click', loadTables);

// 选择表
async function selectTable(tableName) {
    // 更新UI
    document.querySelectorAll('.table-item').forEach(item => {
        item.classList.remove('active');
        if (item.textContent === tableName) {
            item.classList.add('active');
        }
    });
    
    currentTable = tableName;
    currentPage = 1;
    
    // 切换到数据标签页
    switchTab('data');
    await loadTableData();
    await loadTableSchema();
}

// 存储列信息（用于排序）
let currentColumns = [];

// 加载表数据
async function loadTableData() {
    if (!currentTable) return;
    
    try {
        // 先获取列信息，确保按正确顺序显示
        const columnsResponse = await fetch(`/api/table/columns?table=${currentTable}`);
        const columnsData = await columnsResponse.json();
        
        if (columnsData.success) {
            currentColumns = columnsData.columns.map(col => col.name);
        }
        
        // 然后获取数据
        const response = await fetch(`/api/table/data?table=${currentTable}&page=${currentPage}&pageSize=${pageSize}`);
        const data = await response.json();
        
        if (data.success) {
            // 按照 data.columns 的顺序显示数据
            const dataByColumns = [];
            const columns = data.data.columns;            
            data.data.data.forEach(row => {
                const rowByColumns = {};
                columns.forEach(col => {
                    rowByColumns[col.name] = row[col.name];
                });
                dataByColumns.push(rowByColumns);
            });

            displayTableData(dataByColumns, data.total);
            updatePagination(data.total, data.page, data.pageSize);
        }
    } catch (error) {
        showNotification('加载数据失败: ' + error.message, 'error');
    }
}

// 显示表数据
function displayTableData(rows, total) {
    // 清空表格内容，避免DOM操作冲突
    while (dataTableHead.firstChild) {
        dataTableHead.removeChild(dataTableHead.firstChild);
    }
    while (dataTableBody.firstChild) {
        dataTableBody.removeChild(dataTableBody.firstChild);
    }
    
    if (rows.length === 0) {
        const emptyRow = document.createElement('tr');
        const emptyCell = document.createElement('td');
        emptyCell.colSpan = 100;
        emptyCell.style.cssText = 'text-align: center; padding: 2rem; color: var(--text-secondary);';
        emptyCell.textContent = '没有数据';
        emptyRow.appendChild(emptyCell);
        dataTableBody.appendChild(emptyRow);
        return;
    }
    
    // 获取列名，严格按照 currentColumns 的顺序
    let columns;
    if (currentColumns.length > 0) {
        // 使用获取到的列顺序，只包含数据中实际存在的列
        const rowKeys = new Set(Object.keys(rows[0]));
        columns = currentColumns.filter(col => rowKeys.has(col));
        // 添加数据中存在但列信息中不存在的列（以防万一，放在最后）
        Object.keys(rows[0]).forEach(key => {
            if (!columns.includes(key)) {
                columns.push(key);
            }
        });
    } else {
        // 如果没有列信息，使用对象键（降级方案）
        columns = Object.keys(rows[0]);
    }
    
    // 创建表头
    const headRow = document.createElement('tr');
    columns.forEach(col => {
        const th = document.createElement('th');
        th.textContent = col;
        headRow.appendChild(th);
    });
    const actionTh = document.createElement('th');
    actionTh.style.width = '150px';
    actionTh.textContent = '操作';
    headRow.appendChild(actionTh);
    dataTableHead.appendChild(headRow);
    
    // 创建表体
    rows.forEach((row, index) => {
        const bodyRow = document.createElement('tr');
        
        // 按照列顺序添加单元格
        columns.forEach(col => {
            const td = document.createElement('td');
            const value = row[col];
            if (value === null || value === undefined) {
                const nullSpan = document.createElement('span');
                nullSpan.style.color = 'var(--text-secondary)';
                nullSpan.textContent = 'NULL';
                td.appendChild(nullSpan);
            } else {
                td.textContent = String(value);
            }
            bodyRow.appendChild(td);
        });
        
        // 添加操作列
        const actionTd = document.createElement('td');
        const editBtn = document.createElement('button');
        editBtn.className = 'btn btn-secondary action-btn edit-row-btn';
        editBtn.textContent = '编辑';
        editBtn.dataset.row = JSON.stringify(row);
        
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-danger action-btn delete-row-btn';
        deleteBtn.textContent = '删除';
        deleteBtn.dataset.row = JSON.stringify(row);
        
        actionTd.appendChild(editBtn);
        actionTd.appendChild(deleteBtn);
        bodyRow.appendChild(actionTd);
        
        dataTableBody.appendChild(bodyRow);
    });
    
    // 绑定事件监听器
    dataTableBody.querySelectorAll('.edit-row-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const rowData = JSON.parse(this.dataset.row);
            editRow(rowData);
        });
    });
    
    dataTableBody.querySelectorAll('.delete-row-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const rowData = JSON.parse(this.dataset.row);
            deleteRow(rowData);
        });
    });
}

// 更新分页
function updatePagination(total, page, pageSize) {
    const totalPages = Math.ceil(total / pageSize);
    
    paginationInfo.textContent = `共 ${total} 条，第 ${page}/${totalPages} 页`;
    
    let paginationHTML = '';
    paginationHTML += `<button ${page === 1 ? 'disabled' : ''} onclick="changePage(${page - 1})">上一页</button>`;
    
    for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
        paginationHTML += `<button class="${i === page ? 'active' : ''}" onclick="changePage(${i})">${i}</button>`;
    }
    
    paginationHTML += `<button ${page === totalPages ? 'disabled' : ''} onclick="changePage(${page + 1})">下一页</button>`;
    pagination.innerHTML = paginationHTML;
}

// 切换页码
function changePage(page) {
    currentPage = page;
    loadTableData();
}

// 刷新数据
refreshData.addEventListener('click', loadTableData);

// 加载表结构
async function loadTableSchema() {
    if (!currentTable) return;
    
    try {
        const response = await fetch(`/api/table/schema?table=${currentTable}`);
        const data = await response.json();
        
        if (data.success) {
            schemaContent.textContent = data.schema;
        }
    } catch (error) {
        showNotification('加载表结构失败: ' + error.message, 'error');
    }
}

// 标签页切换
tabs.forEach(tab => {
    tab.addEventListener('click', () => {
        const tabName = tab.dataset.tab;
        switchTab(tabName);
    });
});

function switchTab(tabName) {
    tabs.forEach(t => t.classList.remove('active'));
    tabContents.forEach(tc => tc.classList.remove('active'));
    
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    document.getElementById(`${tabName}Tab`).classList.add('active');
    
    if (tabName === 'schema' && currentTable) {
        loadTableSchema();
    }
}

// 执行SQL查询
executeQuery.addEventListener('click', async () => {
    const query = sqlQuery.value.trim();
    if (!query) {
        showNotification('请输入SQL查询', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/query', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ query })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            if (data.data) {
                // 查询结果
                displayQueryResults(data.data);
            } else if (data.affected !== undefined) {
                // 更新/删除/插入结果
                queryResults.innerHTML = `<div class="query-message success">操作成功，影响 ${data.affected} 行</div>`;
            }
        } else {
            queryResults.innerHTML = `<div class="query-message error">${data.message || '执行失败'}</div>`;
        }
    } catch (error) {
        queryResults.innerHTML = `<div class="query-message error">执行失败: ${error.message}</div>`;
    }
});

// 显示查询结果
function displayQueryResults(rows) {
    if (rows.length === 0) {
        queryResults.innerHTML = '<div class="query-message">查询结果为空</div>';
        return;
    }
    
    const columns = Object.keys(rows[0]);
    
    let html = '<table><thead><tr>';
    columns.forEach(col => {
        html += `<th>${escapeHtml(col)}</th>`;
    });
    html += '</tr></thead><tbody>';
    
    rows.forEach(row => {
        html += '<tr>';
        columns.forEach(col => {
            const value = row[col];
            html += `<td>${value === null ? '<span style="color: var(--text-secondary);">NULL</span>' : escapeHtml(String(value))}</td>`;
        });
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    queryResults.innerHTML = html;
}

// 清空查询
clearQuery.addEventListener('click', () => {
    sqlQuery.value = '';
    queryResults.innerHTML = '';
});

// 编辑行（全局函数，供外部调用）
window.editRow = function(rowData) {
    currentRowData = rowData;
    
    // 获取列信息
    fetch(`/api/table/columns?table=${currentTable}`)
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                let formHTML = '';
                data.columns.forEach(col => {
                    const value = rowData[col.name] || '';
                    formHTML += `
                        <div class="edit-form-group">
                            <label>${escapeHtml(col.name)} <span style="color: var(--text-secondary);">(${col.type})</span></label>
                            <input type="text" id="edit_${col.name}" value="${escapeHtml(String(value))}" ${col.key === 'PRI' ? 'readonly style="background: var(--surface);"' : ''}>
                        </div>
                    `;
                });
                editForm.innerHTML = formHTML;
                editModal.style.display = 'flex';
            }
        })
        .catch(err => {
            showNotification('加载列信息失败: ' + err.message, 'error');
        });
}

// 保存编辑
saveEdit.addEventListener('click', async () => {
    if (!currentTable || !currentRowData) return;
    
    // 获取主键列
    const columns = await fetch(`/api/table/columns?table=${currentTable}`)
        .then(res => res.json())
        .then(data => data.columns);
    
    const primaryKeys = columns.filter(col => col.key === 'PRI');
    
    // 构建WHERE条件（使用主键）
    const where = {};
    primaryKeys.forEach(pk => {
        where[pk.name] = currentRowData[pk.name];
    });
    
    // 构建更新数据
    const updateData = {};
    columns.forEach(col => {
        if (col.key !== 'PRI') {
            const input = document.getElementById(`edit_${col.name}`);
            if (input) {
                const value = input.value.trim();
                updateData[col.name] = value === '' ? null : value;
            }
        }
    });
    
    try {
        const response = await fetch('/api/row/update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                table: currentTable,
                data: updateData,
                where: where
            })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showNotification('更新成功', 'success');
            editModal.style.display = 'none';
            loadTableData();
        } else {
            showNotification(data.message || '更新失败', 'error');
        }
    } catch (error) {
        showNotification('更新失败: ' + error.message, 'error');
    }
});

// 删除行（全局函数，供外部调用）
window.deleteRow = function(rowData) {
    currentRowData = rowData;
    
    // 获取主键列
    fetch(`/api/table/columns?table=${currentTable}`)
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                const primaryKeys = data.columns.filter(col => col.key === 'PRI');
                const where = {};
                primaryKeys.forEach(pk => {
                    where[pk.name] = rowData[pk.name];
                });
                currentDeleteWhere = where;
                deleteModal.style.display = 'flex';
            }
        })
        .catch(err => {
            showNotification('加载列信息失败: ' + err.message, 'error');
        });
}

// 确认删除
confirmDelete.addEventListener('click', async () => {
    if (!currentTable || !currentDeleteWhere) return;
    
    try {
        const response = await fetch('/api/row/delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                table: currentTable,
                where: currentDeleteWhere
            })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showNotification('删除成功', 'success');
            deleteModal.style.display = 'none';
            loadTableData();
        } else {
            showNotification(data.message || '删除失败', 'error');
        }
    } catch (error) {
        showNotification('删除失败: ' + error.message, 'error');
    }
});

// 关闭模态框
closeEditModal.addEventListener('click', () => {
    editModal.style.display = 'none';
});

cancelEdit.addEventListener('click', () => {
    editModal.style.display = 'none';
});

closeDeleteModal.addEventListener('click', () => {
    deleteModal.style.display = 'none';
});

cancelDelete.addEventListener('click', () => {
    deleteModal.style.display = 'none';
});

// 工具函数
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showNotification(message, type) {
    // 简单的通知实现
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 1rem 1.5rem;
        background: ${type === 'success' ? 'var(--success-color)' : 'var(--danger-color)'};
        color: white;
        border-radius: 4px;
        box-shadow: var(--shadow);
        z-index: 10000;
        animation: slideIn 0.3s;
    `;
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// 添加动画样式
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

