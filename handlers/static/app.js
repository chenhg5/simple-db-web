// ==================== 全局配置和扩展机制 ====================
// 全局配置对象，允许外部项目自定义行为
window.SimpleDBConfig = window.SimpleDBConfig || {
    // 请求拦截器：在发送请求前可以修改请求配置
    // 参数: (url, options) => { return { url, options }; }
    // options 包含 method, headers, body 等 fetch 标准选项
    requestInterceptor: null,
    
    // 响应拦截器：在收到响应后可以处理响应
    // 参数: (response) => { return response; }
    responseInterceptor: null,
    
    // 错误拦截器：在请求出错时处理错误
    // 参数: (error, url, options) => { return error; }
    errorInterceptor: null
};

// 统一的API请求函数，支持拦截器
async function apiRequest(url, options = {}) {
    // 默认headers
    const defaultHeaders = {};
    
    // 如果有body且是对象或字符串，默认添加Content-Type
    if (options.body) {
        if (typeof options.body === 'string' || (typeof options.body === 'object' && !(options.body instanceof FormData))) {
            defaultHeaders['Content-Type'] = 'application/json';
        }
    }
    
    // 合并headers（用户自定义的headers优先级更高）
    const headers = {
        ...defaultHeaders,
        ...(options.headers || {})
    };
    
    // 添加连接ID到headers（如果存在）
    if (connectionId) {
        headers['X-Connection-ID'] = connectionId;
    }
    
    // 构建请求配置
    let requestOptions = {
        ...options,
        headers: headers
    };
    
    // 调用请求拦截器（如果存在）
    if (window.SimpleDBConfig.requestInterceptor) {
        try {
            const intercepted = window.SimpleDBConfig.requestInterceptor(url, requestOptions);
            if (intercepted) {
                url = intercepted.url || url;
                requestOptions = intercepted.options || requestOptions;
            }
        } catch (error) {
            console.warn('请求拦截器执行失败:', error);
        }
    }
    
    try {
        // 发送请求
        let response = await fetch(url, requestOptions);
        
        // 调用响应拦截器（如果存在）
        if (window.SimpleDBConfig.responseInterceptor) {
            try {
                response = await window.SimpleDBConfig.responseInterceptor(response);
            } catch (error) {
                console.warn('响应拦截器执行失败:', error);
            }
        }
        
        return response;
    } catch (error) {
        // 调用错误拦截器（如果存在）
        if (window.SimpleDBConfig.errorInterceptor) {
            try {
                error = await window.SimpleDBConfig.errorInterceptor(error, url, requestOptions);
            } catch (interceptorError) {
                console.warn('错误拦截器执行失败:', interceptorError);
            }
        }
        throw error;
    }
}

// 导出配置对象和请求函数到全局，方便外部访问
window.SimpleDB = window.SimpleDB || {};
window.SimpleDB.config = window.SimpleDBConfig;
window.SimpleDB.apiRequest = apiRequest;

// ==================== 全局状态 ====================
let currentTable = null;
let currentPage = 1;
let pageSize = 50;
let currentRowData = null;
let currentDeleteWhere = null;
let connectionId = null; // 当前连接的ID
let connectionInfo = null; // 当前连接信息
let currentDbType = null; // 当前数据库类型

// API 基础路径，动态获取以支持路由前缀
// 获取当前页面的基础路径（去掉文件名，保留路径部分）
function getBasePath() {
    const path = window.location.pathname;
    // 去掉末尾的斜杠（如果有）
    const basePath = path.endsWith('/') ? path.slice(0, -1) : path;
    // 如果路径为空，返回空字符串（根路径）
    return basePath || '';
}

// API 基础路径
const API_BASE = `${getBasePath()}/api`;

// DOM元素
const connectionStatus = document.getElementById('connectionStatus');
const connectionInfoElement = document.getElementById('connectionInfo');
const connectionInfoText = document.getElementById('connectionInfoText');
const connectionForm = document.getElementById('connectionForm');
const connectionMode = document.getElementById('connectionMode');
const dsnGroup = document.getElementById('dsnGroup');
const formGroup = document.getElementById('formGroup');
const connectionsPanel = document.getElementById('connectionsPanel');
const activeConnectionsList = document.getElementById('activeConnectionsList');
const newConnectionBtn = document.getElementById('newConnectionBtn');
const newConnectionModal = document.getElementById('newConnectionModal');
const closeNewConnectionModal = document.getElementById('closeNewConnectionModal');
const cancelNewConnection = document.getElementById('cancelNewConnection');
const confirmNewConnection = document.getElementById('confirmNewConnection');
const useProxy = document.getElementById('useProxy');
const proxyGroup = document.getElementById('proxyGroup');
const proxyType = document.getElementById('proxyType');
const proxyHost = document.getElementById('proxyHost');
const proxyPort = document.getElementById('proxyPort');
const proxyUser = document.getElementById('proxyUser');
const proxyPassword = document.getElementById('proxyPassword');
const proxyKeyData = document.getElementById('proxyKeyData');
const toggleProxyPassword = document.getElementById('toggleProxyPassword');
const databasePanel = document.getElementById('databasePanel');
const databaseSelect = document.getElementById('databaseSelect');
const disconnectBtn = document.getElementById('disconnectBtn');
const tablesPanel = document.getElementById('tablesPanel');
const tableList = document.getElementById('tableList');
const refreshTables = document.getElementById('refreshTables');
const tableFilter = document.getElementById('tableFilter');
const rememberConnection = document.getElementById('rememberConnection');
const savedConnectionsPanel = document.getElementById('savedConnectionsPanel');
const savedConnectionsList = document.getElementById('savedConnectionsList');
const clearSavedConnections = document.getElementById('clearSavedConnections');
// Loading 元素
const dataLoading = document.getElementById('dataLoading');
const schemaLoading = document.getElementById('schemaLoading');
const queryLoading = document.getElementById('queryLoading');
const databaseLoading = document.getElementById('databaseLoading');
const tablesLoading = document.getElementById('tablesLoading');
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
const pageSizeSelect = document.getElementById('pageSizeSelect');
const schemaContent = document.getElementById('schemaContent');
const copySchemaBtn = document.getElementById('copySchemaBtn');
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
const togglePassword = document.getElementById('togglePassword');
const deleteConnectionModal = document.getElementById('deleteConnectionModal');
const closeDeleteConnectionModal = document.getElementById('closeDeleteConnectionModal');
const cancelDeleteConnection = document.getElementById('cancelDeleteConnection');
const confirmDeleteConnection = document.getElementById('confirmDeleteConnection');
const clearAllConnectionsModal = document.getElementById('clearAllConnectionsModal');
const closeClearAllConnectionsModal = document.getElementById('closeClearAllConnectionsModal');
const cancelClearAllConnections = document.getElementById('cancelClearAllConnections');
const confirmClearAllConnections = document.getElementById('confirmClearAllConnections');

// 删除连接相关的状态
let deleteConnectionIndex = null;

// 活动连接列表（支持多个连接）
let activeConnections = new Map(); // connectionId -> connectionInfo

// 密码显示/隐藏切换
if (togglePassword) {
    togglePassword.addEventListener('click', () => {
        const passwordInput = document.getElementById('password');
        if (passwordInput.type === 'password') {
            passwordInput.type = 'text';
            togglePassword.textContent = '🙈';
        } else {
            passwordInput.type = 'password';
            togglePassword.textContent = '👁️';
        }
    });
}

// 代理密码显示/隐藏切换
if (toggleProxyPassword) {
    toggleProxyPassword.addEventListener('click', () => {
        if (proxyPassword.type === 'password') {
            proxyPassword.type = 'text';
            toggleProxyPassword.textContent = '🙈';
        } else {
            proxyPassword.type = 'password';
            toggleProxyPassword.textContent = '👁️';
        }
    });
}

// 代理配置显示/隐藏
if (useProxy) {
    useProxy.addEventListener('change', (e) => {
        if (e.target.checked) {
            proxyGroup.style.display = 'block';
        } else {
            proxyGroup.style.display = 'none';
        }
    });
}

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

// 密码加密/解密函数（简单的 Base64 编码，不是真正的加密，但至少不是明文）
function encryptPassword(password) {
    if (!password) return '';
    return btoa(unescape(encodeURIComponent(password)));
}

function decryptPassword(encrypted) {
    if (!encrypted) return '';
    try {
        return decodeURIComponent(escape(atob(encrypted)));
    } catch (e) {
        return '';
    }
}

// 生成连接的唯一标识（用于去重）
function getConnectionKey(connectionInfo) {
    if (connectionInfo.dsn) {
        // DSN 模式：提取 host、port、user
        const dsn = connectionInfo.dsn;
        const userMatch = dsn.match(/^([^:]+):/);
        const hostMatch = dsn.match(/@tcp\(([^:]+)/);
        const portMatch = dsn.match(/@tcp\([^:]+:(\d+)/);
        const user = userMatch ? userMatch[1] : '';
        const host = hostMatch ? hostMatch[1] : '';
        const port = portMatch ? portMatch[1] : '3306';
        return `${connectionInfo.type}:${host}:${port}:${user}`;
    } else {
        return `${connectionInfo.type}:${connectionInfo.host || ''}:${connectionInfo.port || '3306'}:${connectionInfo.user || ''}`;
    }
}

// 保存连接信息到 localStorage
function saveConnection(connectionInfo) {
    try {
        const saved = getSavedConnections();
        const key = getConnectionKey(connectionInfo);
        
        // 检查是否已存在（去重）
        const existingIndex = saved.findIndex(conn => getConnectionKey(conn) === key);
        
        const connectionToSave = {
            ...connectionInfo,
            savedAt: new Date().toISOString()
        };
        
        // 如果使用表单模式，加密密码
        if (!connectionToSave.dsn && connectionToSave.password) {
            connectionToSave.password = encryptPassword(connectionToSave.password);
            connectionToSave.passwordEncrypted = true;
        }
        
        if (existingIndex >= 0) {
            // 更新已存在的连接
            const existingConn = saved[existingIndex];
            // 如果新连接没有密码字段，保留旧的密码和 passwordEncrypted 字段
            if (!connectionToSave.password && existingConn.password) {
                connectionToSave.password = existingConn.password;
                connectionToSave.passwordEncrypted = existingConn.passwordEncrypted;
            }
            saved[existingIndex] = connectionToSave;
        } else {
            // 添加新连接
            saved.push(connectionToSave);
        }
        
        localStorage.setItem('savedConnections', JSON.stringify(saved));
        loadSavedConnections();
    } catch (error) {
        console.error('保存连接失败:', error);
    }
}

// 从 localStorage 获取保存的连接
function getSavedConnections() {
    try {
        const saved = localStorage.getItem('savedConnections');
        return saved ? JSON.parse(saved) : [];
    } catch (error) {
        console.error('读取保存的连接失败:', error);
        return [];
    }
}

// 加载并显示保存的连接
function loadSavedConnections() {
    const saved = getSavedConnections();
    savedConnectionsList.innerHTML = '';
    
    if (saved.length === 0) {
        const emptyMsg = document.createElement('div');
        emptyMsg.style.cssText = 'padding: 1rem; color: var(--text-secondary); text-align: center; font-size: 0.875rem;';
        emptyMsg.textContent = '暂无保存的连接';
        savedConnectionsList.appendChild(emptyMsg);
        return;
    }
    
    saved.forEach((conn, index) => {
        let displayText = '';
        if (conn.dsn) {
            // DSN 模式
            const userMatch = conn.dsn.match(/^([^:]+):/);
            const hostMatch = conn.dsn.match(/@tcp\(([^:]+)/);
            const user = userMatch ? userMatch[1] : 'unknown';
            const host = hostMatch ? hostMatch[1] : 'unknown';
            displayText = `${conn.type || 'mysql'}://${user}@${host}`;
        } else {
            displayText = `${conn.type || 'mysql'}://${conn.user || 'unknown'}@${conn.host || 'unknown'}:${conn.port || '3306'}`;
        }
        
        // 创建按钮容器
        const buttonWrapper = document.createElement('div');
        buttonWrapper.style.cssText = 'margin-bottom: 0.5rem; display: flex; align-items: center; gap: 0.5rem;';
        
        // 创建连接按钮
        const connectBtn = document.createElement('button');
        connectBtn.className = 'btn btn-secondary';
        connectBtn.style.cssText = 'flex: 1; text-align: left; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding: 0.5rem 0.75rem; font-size: 0.875rem;';
        connectBtn.textContent = displayText;
        connectBtn.title = displayText; // 完整文本作为提示
        
        // 创建删除按钮
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-secondary';
        deleteBtn.style.cssText = 'flex-shrink: 0; width: 2rem; padding: 0.5rem; font-size: 0.875rem; line-height: 1;';
        deleteBtn.textContent = '×';
        deleteBtn.title = '删除';
        deleteBtn.dataset.index = index;
        
        // 点击连接按钮
        connectBtn.addEventListener('click', () => {
            connectWithSavedConnection(conn);
        });
        
        // 点击删除按钮
        deleteBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            deleteConnectionIndex = index;
            deleteConnectionModal.style.display = 'flex';
        });
        
        buttonWrapper.appendChild(connectBtn);
        buttonWrapper.appendChild(deleteBtn);
        savedConnectionsList.appendChild(buttonWrapper);
    });
}

// 使用保存的连接进行连接
async function connectWithSavedConnection(savedConn) {
    // 填充表单
    document.getElementById('dbType').value = savedConn.type || 'mysql';
    
    let connectionInfo = {
        type: savedConn.type || 'mysql'
    };
    
    if (savedConn.dsn) {
        // DSN 模式
        connectionMode.value = 'dsn';
        document.getElementById('dsn').value = savedConn.dsn;
        dsnGroup.style.display = 'block';
        formGroup.style.display = 'none';
        connectionInfo.dsn = savedConn.dsn;
    } else {
        // 表单模式
        connectionMode.value = 'form';
        document.getElementById('host').value = savedConn.host || '';
        document.getElementById('port').value = savedConn.port || '3306';
        document.getElementById('user').value = savedConn.user || '';
        
        // 解密密码
        let password = '';
        if (savedConn.passwordEncrypted) {
            password = decryptPassword(savedConn.password);
        } else {
            password = savedConn.password || '';
        }
        document.getElementById('password').value = password;
        
        connectionInfo.host = savedConn.host || '';
        connectionInfo.port = savedConn.port || '3306';
        connectionInfo.user = savedConn.user || '';
        connectionInfo.password = password;
        connectionInfo.database = '';
        
        dsnGroup.style.display = 'none';
        formGroup.style.display = 'block';
    }
    
    // 直接执行连接逻辑，避免重复提交
    const connectBtn = connectionForm.querySelector('button[type="submit"]');
    setButtonLoading(connectBtn, true);
    try {
        const response = await apiRequest(`${API_BASE}/connect`, {
            method: 'POST',
            body: JSON.stringify(connectionInfo)
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            // 保存连接ID和连接信息
            connectionId = data.connectionId;
            const connInfo = {
                type: savedConn.type || 'mysql',
                host: savedConn.host || '',
                port: savedConn.port || '3306',
                user: savedConn.user || '',
                dsn: savedConn.dsn || ''
            };
            connectionInfo = connInfo;
            currentDbType = savedConn.type || 'mysql'; // 保存数据库类型
            sessionStorage.setItem('currentConnectionId', connectionId);
            sessionStorage.setItem('currentConnectionInfo', JSON.stringify(connInfo));
            updateConnectionStatus(true);
            updateConnectionInfo(connInfo);
            // 检查DSN中是否包含数据库
            const dsn = connInfo.dsn || '';
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
    } finally {
        setButtonLoading(connectBtn, false);
    }
}

// 删除保存的连接
function deleteSavedConnection(index) {
    const saved = getSavedConnections();
    saved.splice(index, 1);
    localStorage.setItem('savedConnections', JSON.stringify(saved));
    loadSavedConnections();
}

// 确认删除连接
confirmDeleteConnection.addEventListener('click', () => {
    if (deleteConnectionIndex !== null) {
        deleteSavedConnection(deleteConnectionIndex);
        deleteConnectionModal.style.display = 'none';
        deleteConnectionIndex = null;
        showNotification('已删除连接', 'success');
    }
});

// 取消删除连接
cancelDeleteConnection.addEventListener('click', () => {
    deleteConnectionModal.style.display = 'none';
    deleteConnectionIndex = null;
});

closeDeleteConnectionModal.addEventListener('click', () => {
    deleteConnectionModal.style.display = 'none';
    deleteConnectionIndex = null;
});

// 清空所有保存的连接
clearSavedConnections.addEventListener('click', () => {
    clearAllConnectionsModal.style.display = 'flex';
});

// 确认清除所有连接
confirmClearAllConnections.addEventListener('click', () => {
    localStorage.removeItem('savedConnections');
    loadSavedConnections();
    clearAllConnectionsModal.style.display = 'none';
    showNotification('已清空所有保存的连接', 'success');
});

// 取消清除所有连接
cancelClearAllConnections.addEventListener('click', () => {
    clearAllConnectionsModal.style.display = 'none';
});

closeClearAllConnectionsModal.addEventListener('click', () => {
    clearAllConnectionsModal.style.display = 'none';
});

// 存储数据库类型列表
let databaseTypes = [];

// Loading 控制函数
function showLoading(element) {
    if (element) {
        element.style.display = 'flex';
    }
}

function hideLoading(element) {
    if (element) {
        element.style.display = 'none';
    }
}

function setButtonLoading(button, loading) {
    if (!button) return;
    if (loading) {
        button.classList.add('loading');
        button.disabled = true;
    } else {
        button.classList.remove('loading');
        button.disabled = false;
    }
}

// 加载数据库类型列表
async function loadDatabaseTypes() {
    try {
        const response = await apiRequest(`${API_BASE}/database/types`);
        const data = await response.json();
        
        if (data.success && data.types) {
            databaseTypes = data.types;
            updateDatabaseTypeSelect();
        }
    } catch (error) {
        console.error('加载数据库类型失败:', error);
        // 如果加载失败，使用默认类型
        databaseTypes = [
            { type: 'mysql', display_name: 'MySQL' },
            { type: 'postgresql', display_name: 'PostgreSQL' },
            { type: 'sqlite', display_name: 'SQLite' }
        ];
        updateDatabaseTypeSelect();
    }
}

// 更新数据库类型选择框
function updateDatabaseTypeSelect() {
    const dbTypeSelect = document.getElementById('dbType');
    if (!dbTypeSelect) return;
    
    // 清空现有选项
    dbTypeSelect.innerHTML = '';
    
    // 添加数据库类型选项
    databaseTypes.forEach(dbType => {
        const option = document.createElement('option');
        option.value = dbType.type;
        option.textContent = dbType.display_name;
        dbTypeSelect.appendChild(option);
    });
}

// 页面加载时加载保存的连接
loadSavedConnections();

// 页面加载时加载数据库类型列表
loadDatabaseTypes();

// 页面加载时尝试恢复连接
async function restoreConnection() {
    try {
        // 从 sessionStorage 获取保存的连接ID和连接信息
        const savedConnectionId = sessionStorage.getItem('currentConnectionId');
        const savedConnectionInfo = sessionStorage.getItem('currentConnectionInfo');
        
        if (!savedConnectionId) {
            return;
        }
        
        // 检查连接是否仍然有效
        // 临时设置connectionId以便apiRequest自动添加header
        const originalConnectionId = connectionId;
        connectionId = savedConnectionId;
        const response = await apiRequest(`${API_BASE}/status`, {
            headers: {
                'X-Connection-ID': savedConnectionId
            }
        });
        connectionId = originalConnectionId;
        const data = await response.json();
        
        if (response.ok && data.connected) {
            // 恢复连接ID和连接信息
            connectionId = savedConnectionId;
            if (savedConnectionInfo) {
                connectionInfo = JSON.parse(savedConnectionInfo);
                currentDbType = data.dbType || connectionInfo.type || null; // 恢复数据库类型
                
                // 添加到活动连接列表
                activeConnections.set(savedConnectionId, {
                    connectionId: savedConnectionId,
                    connectionInfo: connectionInfo,
                    databases: data.databases || []
                });
                
                updateConnectionInfo(connectionInfo);
            }
            // 有活动的连接，恢复UI状态
            updateConnectionStatus(true);
            updateActiveConnectionsList();
            databasePanel.style.display = 'block';
            
            // 加载数据库列表
            await loadDatabases(data.databases || []);
            
            // 如果有当前数据库，恢复它
            if (data.currentDatabase) {
                databaseSelect.value = data.currentDatabase;
                await switchDatabase(data.currentDatabase);
            }
            
            // 如果有当前表，恢复它
            if (data.currentTable) {
                currentTable = data.currentTable;
                await loadTableData();
                await loadTableSchema();
            }
        } else {
            // 连接已失效，清除保存的连接ID
            sessionStorage.removeItem('currentConnectionId');
            sessionStorage.removeItem('currentConnectionInfo');
            connectionId = null;
            connectionInfo = null;
        }
    } catch (error) {
        // 连接失败，保持未连接状态
        console.log('无法恢复连接:', error);
        connectionId = null;
        connectionInfo = null;
        sessionStorage.removeItem('currentConnectionId');
        sessionStorage.removeItem('currentConnectionInfo');
    }
}

// 页面加载完成后尝试恢复连接
document.addEventListener('DOMContentLoaded', () => {
    restoreConnection();
});

// 新增连接按钮点击事件
if (newConnectionBtn) {
    newConnectionBtn.addEventListener('click', () => {
        // 清空表单
        if (connectionForm) {
            connectionForm.reset();
        }
        // 重置代理配置
        if (useProxy) {
            useProxy.checked = false;
            proxyGroup.style.display = 'none';
        }
        // 显示模态框
        if (newConnectionModal) {
            newConnectionModal.style.display = 'flex';
        }
    });
}

// 关闭新增连接模态框
if (closeNewConnectionModal) {
    closeNewConnectionModal.addEventListener('click', () => {
        if (newConnectionModal) {
            newConnectionModal.style.display = 'none';
        }
    });
}

if (cancelNewConnection) {
    cancelNewConnection.addEventListener('click', () => {
        if (newConnectionModal) {
            newConnectionModal.style.display = 'none';
        }
    });
}

// 连接数据库（在模态框中）
if (confirmNewConnection) {
    confirmNewConnection.addEventListener('click', async () => {
        await handleConnect();
    });
}

// 连接表单提交（兼容旧代码）
if (connectionForm) {
    connectionForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleConnect();
    });
}

// 统一的连接处理函数
async function handleConnect() {
    const mode = connectionMode ? connectionMode.value : 'form';
    const dbType = document.getElementById('dbType') ? document.getElementById('dbType').value : '';
    
    if (!dbType) {
        showNotification('请选择数据库类型', 'error');
        return;
    }
    
    let connectionInfo = {
        type: dbType
    };
    
    // 构建连接信息
    if (mode === 'dsn') {
        const dsnInput = document.getElementById('dsn');
        if (dsnInput && dsnInput.value) {
            connectionInfo.dsn = dsnInput.value;
        } else {
            showNotification('请输入DSN连接字符串', 'error');
            return;
        }
    } else {
        const hostInput = document.getElementById('host');
        const userInput = document.getElementById('user');
        if (!hostInput || !hostInput.value || !userInput || !userInput.value) {
            showNotification('请填写主机和用户名', 'error');
            return;
        }
        connectionInfo.host = hostInput.value;
        connectionInfo.port = document.getElementById('port') ? (document.getElementById('port').value || '3306') : '3306';
        connectionInfo.user = userInput.value;
        connectionInfo.password = document.getElementById('password') ? document.getElementById('password').value : '';
        connectionInfo.database = '';
    }
    
    // 构建代理配置（如果启用）
    if (useProxy && useProxy.checked) {
        const proxyConfig = {
            type: proxyType ? proxyType.value : 'ssh',
            host: proxyHost ? proxyHost.value : '',
            port: proxyPort ? (proxyPort.value || '22') : '22',
            user: proxyUser ? proxyUser.value : '',
            password: proxyPassword ? proxyPassword.value : '',
            key_file: '',
            config: ''
        };
        
        // 如果提供了SSH私钥
        if (proxyKeyData && proxyKeyData.value) {
            proxyConfig.config = JSON.stringify({
                key_data: proxyKeyData.value
            });
        }
        
        if (!proxyConfig.host || !proxyConfig.user) {
            showNotification('请填写代理主机和用户名', 'error');
            return;
        }
        
        connectionInfo.proxy = proxyConfig;
    }
    
    // 设置按钮加载状态
    const connectBtn = confirmNewConnection || connectionForm?.querySelector('button[type="submit"]');
    if (connectBtn) {
        setButtonLoading(connectBtn, true);
    }
    
    try {
        const response = await apiRequest(`${API_BASE}/connect`, {
            method: 'POST',
            body: JSON.stringify(connectionInfo)
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            // 保存连接ID和连接信息
            const newConnectionId = data.connectionId;
            const connInfo = {
                type: dbType,
                host: mode === 'form' ? (document.getElementById('host')?.value || '') : '',
                port: mode === 'form' ? (document.getElementById('port')?.value || '3306') : '3306',
                user: mode === 'form' ? (document.getElementById('user')?.value || '') : '',
                dsn: mode === 'dsn' ? (document.getElementById('dsn')?.value || '') : '',
                proxy: connectionInfo.proxy || null
            };
            
            // 添加到活动连接列表
            activeConnections.set(newConnectionId, {
                connectionId: newConnectionId,
                connectionInfo: connInfo,
                databases: data.databases || []
            });
            
            // 更新当前连接（兼容旧代码）
            connectionId = newConnectionId;
            connectionInfo = connInfo;
            currentDbType = dbType;
            
            // 保存到sessionStorage（用于页面刷新后恢复）
            sessionStorage.setItem('currentConnectionId', newConnectionId);
            sessionStorage.setItem('currentConnectionInfo', JSON.stringify(connInfo));
            
            // 更新UI
            updateConnectionStatus(true);
            updateConnectionInfo(connInfo);
            updateActiveConnectionsList();
            
            // 如果勾选了"记住连接"，保存连接信息
            if (rememberConnection && rememberConnection.checked) {
                const connectionToSave = {
                    ...connInfo,
                    password: mode === 'form' ? (document.getElementById('password')?.value || '') : ''
                };
                saveConnection(connectionToSave);
            }
            
            // 关闭模态框
            if (newConnectionModal) {
                newConnectionModal.style.display = 'none';
            }
            
            // 检查DSN中是否包含数据库
            const dsn = mode === 'dsn' ? (document.getElementById('dsn')?.value || '') : '';
            const hasDatabaseInDSN = dsn && (dsn.includes('/') && !dsn.endsWith('/') && !dsn.endsWith('/?'));
            
            if (hasDatabaseInDSN) {
                // DSN中包含数据库,直接使用该数据库
                databasePanel.style.display = 'block';
                await loadDatabases(data.databases || []);
                const dbMatch = dsn.match(/\/([^\/\?]+)/);
                if (dbMatch && dbMatch[1]) {
                    const dbName = dbMatch[1];
                    databaseSelect.value = dbName;
                    await switchDatabase(dbName);
                } else {
                    await loadTables();
                }
            } else {
                // DSN中不包含数据库,显示数据库选择器
                databasePanel.style.display = 'block';
                await loadDatabases(data.databases || []);
            }
            showNotification('连接成功', 'success');
        } else {
            showNotification(data.message || '连接失败', 'error');
        }
    } catch (error) {
        showNotification('连接失败: ' + error.message, 'error');
    } finally {
        if (connectBtn) {
            setButtonLoading(connectBtn, false);
        }
    }
}

// 更新活动连接列表
function updateActiveConnectionsList() {
    if (!activeConnectionsList) return;
    
    activeConnectionsList.innerHTML = '';
    
    if (activeConnections.size === 0) {
        const emptyMsg = document.createElement('div');
        emptyMsg.style.cssText = 'padding: 1rem; color: var(--text-secondary); text-align: center; font-size: 0.875rem;';
        emptyMsg.textContent = '暂无活动连接';
        activeConnectionsList.appendChild(emptyMsg);
        return;
    }
    
    activeConnections.forEach((conn, connId) => {
        const connItem = document.createElement('div');
        connItem.style.cssText = 'padding: 0.75rem; margin-bottom: 0.5rem; background: var(--surface); border-radius: 4px; border: 1px solid var(--border-color);';
        
        const info = conn.connectionInfo;
        let displayText = '';
        if (info.dsn) {
            const userMatch = info.dsn.match(/^([^:]+):/);
            const hostMatch = info.dsn.match(/@tcp\(([^:]+)/);
            const user = userMatch ? userMatch[1] : 'unknown';
            const host = hostMatch ? hostMatch[1] : 'unknown';
            displayText = `${info.type || 'mysql'}://${user}@${host}`;
        } else {
            displayText = `${info.type || 'mysql'}://${info.user || 'unknown'}@${info.host || 'unknown'}:${info.port || '3306'}`;
        }
        
        if (info.proxy) {
            displayText += ` [通过${info.proxy.type || 'proxy'}]`;
        }
        
        connItem.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <div style="flex: 1; overflow: hidden;">
                    <div style="font-weight: 600; font-size: 0.875rem; margin-bottom: 0.25rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${displayText}">${displayText}</div>
                    <div style="font-size: 0.75rem; color: var(--text-secondary);">连接ID: ${connId.substring(0, 8)}...</div>
                </div>
                <div style="display: flex; gap: 0.5rem;">
                    <button class="btn btn-secondary switch-connection-btn" data-connection-id="${connId}" style="font-size: 0.75rem; padding: 0.25rem 0.5rem;">切换</button>
                    <button class="btn btn-danger disconnect-connection-btn" data-connection-id="${connId}" style="font-size: 0.75rem; padding: 0.25rem 0.5rem;">断开</button>
                </div>
            </div>
        `;
        
        // 切换连接
        const switchBtn = connItem.querySelector('.switch-connection-btn');
        switchBtn.addEventListener('click', async () => {
            await switchToConnection(connId);
        });
        
        // 断开连接
        const disconnectBtn = connItem.querySelector('.disconnect-connection-btn');
        disconnectBtn.addEventListener('click', async () => {
            await disconnectConnection(connId);
        });
        
        activeConnectionsList.appendChild(connItem);
    });
}

// 切换到指定连接
async function switchToConnection(targetConnectionId) {
    if (!targetConnectionId || !activeConnections.has(targetConnectionId)) {
        showNotification('连接不存在', 'error');
        return;
    }
    
    const conn = activeConnections.get(targetConnectionId);
    connectionId = targetConnectionId;
    connectionInfo = conn.connectionInfo;
    currentDbType = conn.connectionInfo.type;
    
    // 更新sessionStorage
    sessionStorage.setItem('currentConnectionId', targetConnectionId);
    sessionStorage.setItem('currentConnectionInfo', JSON.stringify(conn.connectionInfo));
    
    // 更新UI
    updateConnectionStatus(true);
    updateConnectionInfo(conn.connectionInfo);
    
    // 加载数据库列表
    databasePanel.style.display = 'block';
    await loadDatabases(conn.databases || []);
    
    showNotification('已切换到连接', 'success');
}

// 断开指定连接
async function disconnectConnection(targetConnectionId) {
    if (!targetConnectionId) return;
    
    setButtonLoading(disconnectBtn, true);
    try {
        const response = await apiRequest(`${API_BASE}/disconnect`, {
            method: 'POST',
            headers: {
                'X-Connection-ID': targetConnectionId
            }
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            // 从活动连接列表移除
            activeConnections.delete(targetConnectionId);
            
            // 如果断开的是当前连接，清除当前连接状态
            if (targetConnectionId === connectionId) {
                connectionId = null;
                connectionInfo = null;
                sessionStorage.removeItem('currentConnectionId');
                sessionStorage.removeItem('currentConnectionInfo');
                updateConnectionStatus(false);
                updateConnectionInfo(null);
                databasePanel.style.display = 'none';
                tablesPanel.style.display = 'none';
                currentTable = null;
                databaseSelect.innerHTML = '<option value="">请选择数据库...</option>';
                tableFilter.value = '';
                allTables = [];
                currentColumns = [];
            }
            
            updateActiveConnectionsList();
            showNotification('已断开连接', 'success');
        } else {
            showNotification(data.message || '断开连接失败', 'error');
        }
    } catch (error) {
        showNotification('断开连接失败: ' + error.message, 'error');
    } finally {
        setButtonLoading(disconnectBtn, false);
    }
}

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

// 更新连接信息显示
function updateConnectionInfo(info) {
    if (!info) {
        connectionInfoElement.style.display = 'none';
        return;
    }
    
    let infoText = '';
    // 从数据库类型列表中查找显示名称
    let dbTypeName = info.type;
    if (databaseTypes.length > 0) {
        const dbType = databaseTypes.find(t => t.type === info.type);
        if (dbType) {
            dbTypeName = dbType.display_name;
        }
    } else {
        // 如果列表未加载，使用默认映射
        const dbTypeNames = {
            'mysql': 'MySQL',
            'postgres': 'PostgreSQL',
            'postgresql': 'PostgreSQL',
            'sqlite': 'SQLite',
            'dameng': '达梦',
            'openguass': 'OpenGauss',
            'vastbase': 'Vastbase',
            'kingbase': '人大金仓',
            'oceandb': 'OceanDB'
        };
        dbTypeName = dbTypeNames[info.type] || info.type;
    }
    
    if (info.dsn) {
        // DSN 模式：尝试从 DSN 中提取信息
        const userMatch = info.dsn.match(/^([^:]+):/);
        const hostMatch = info.dsn.match(/@tcp\(([^:]+)/);
        const portMatch = info.dsn.match(/@tcp\([^:]+:(\d+)/);
        const user = userMatch ? userMatch[1] : 'unknown';
        const host = hostMatch ? hostMatch[1] : 'unknown';
        const port = portMatch ? portMatch[1] : '3306';
        infoText = `${dbTypeName}://${user}@${host}:${port}`;
    } else {
        // 表单模式
        const host = info.host || 'localhost';
        const port = info.port || '3306';
        const user = info.user || 'unknown';
        infoText = `${dbTypeName}://${user}@${host}:${port}`;
    }
    
    connectionInfoText.textContent = infoText;
    connectionInfoElement.style.display = 'block';
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
        showLoading(databaseLoading);
        try {
            const response = await apiRequest(`${API_BASE}/databases`);
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
        } finally {
            hideLoading(databaseLoading);
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
    
    showLoading(tablesLoading);
    setButtonLoading(databaseSelect, true);
    try {
        const response = await apiRequest(`${API_BASE}/database/switch`, {
            method: 'POST',
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
    } finally {
        hideLoading(tablesLoading);
        setButtonLoading(databaseSelect, false);
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

// 断开当前连接
if (disconnectBtn) {
    disconnectBtn.addEventListener('click', async () => {
        if (!connectionId) {
            showNotification('没有活动连接', 'error');
            return;
        }
        await disconnectConnection(connectionId);
    });
}

// 加载表列表
async function loadTables() {
    showLoading(tablesLoading);
    setButtonLoading(refreshTables, true);
    try {
        const response = await apiRequest(`${API_BASE}/tables`);
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            showNotification(data.message || '加载表列表失败', 'error');
            hideLoading(tablesLoading);
            setButtonLoading(refreshTables, false);
            return;
        }
        
        if (data.success) {
            displayTables(data.tables || []);
        }
    } catch (error) {
        showNotification('加载表列表失败: ' + error.message, 'error');
    } finally {
        hideLoading(tablesLoading);
        setButtonLoading(refreshTables, false);
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
    // 并行加载数据和结构
    await Promise.all([
        loadTableData(),
        loadTableSchema()
    ]);
}

// 存储列信息（用于排序）
let currentColumns = [];

// 加载表数据
async function loadTableData() {
    if (!currentTable) return;
    
    showLoading(dataLoading);
    setButtonLoading(refreshData, true);
    try {
        // 先获取列信息，确保按正确顺序显示
        const columnsResponse = await apiRequest(`${API_BASE}/table/columns?table=${currentTable}`);
        const columnsData = await columnsResponse.json();
        
        if (!columnsResponse.ok || !columnsData.success) {
            showNotification(columnsData.message || '获取列信息失败', 'error');
            hideLoading(dataLoading);
            setButtonLoading(refreshData, false);
            return;
        }
        
        if (columnsData.success) {
            currentColumns = columnsData.columns.map(col => col.name);
        }
        
        // 然后获取数据
        const response = await apiRequest(`${API_BASE}/table/data?table=${currentTable}&page=${currentPage}&pageSize=${pageSize}`);
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            showNotification(data.message || '获取数据失败', 'error');
            // 即使获取数据失败，如果有列信息，也要显示表头
            if (currentColumns.length > 0) {
                displayTableData([], 0, false);
            }
            hideLoading(dataLoading);
            setButtonLoading(refreshData, false);
            return;
        }
        
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

            // 检查是否为 ClickHouse
            const isClickHouse = data.isClickHouse || false;
            displayTableData(dataByColumns, data.total, isClickHouse);
            updatePagination(data.total, data.page, data.pageSize, isClickHouse);
        }
    } catch (error) {
        showNotification('加载数据失败: ' + error.message, 'error');
    } finally {
        hideLoading(dataLoading);
        setButtonLoading(refreshData, false);
    }
}

// 显示表数据
function displayTableData(rows, total, isClickHouse = false) {
    // 清空表格内容，避免DOM操作冲突
    while (dataTableHead.firstChild) {
        dataTableHead.removeChild(dataTableHead.firstChild);
    }
    while (dataTableBody.firstChild) {
        dataTableBody.removeChild(dataTableBody.firstChild);
    }
    
    // 获取列名，严格按照 currentColumns 的顺序
    let columns;
    if (rows.length > 0) {
        // 有数据时，使用数据中的列
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
    } else {
        // 没有数据时，使用 currentColumns（如果存在）
        columns = currentColumns.length > 0 ? currentColumns : [];
    }
    
    // 创建表头（即使没有数据也要显示表头）
    if (columns.length > 0) {
        const headRow = document.createElement('tr');
        columns.forEach(col => {
            const th = document.createElement('th');
            th.textContent = col;
            headRow.appendChild(th);
        });
        // ClickHouse 不显示操作列
        if (!isClickHouse) {
            const actionTh = document.createElement('th');
            actionTh.className = 'action-column-header';
            actionTh.textContent = '操作';
            headRow.appendChild(actionTh);
        }
        dataTableHead.appendChild(headRow);
    }
    
    // 如果没有数据，显示"没有数据"提示
    if (rows.length === 0) {
        const emptyRow = document.createElement('tr');
        const emptyCell = document.createElement('td');
        const colSpan = columns.length + (isClickHouse ? 0 : 1); // 包括操作列
        emptyCell.colSpan = colSpan;
        emptyCell.style.cssText = 'text-align: center; padding: 2rem; color: var(--text-secondary);';
        emptyCell.textContent = '没有数据';
        emptyRow.appendChild(emptyCell);
        dataTableBody.appendChild(emptyRow);
        return;
    }
    
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
        
        // ClickHouse 不显示操作列
        if (!isClickHouse) {
            const actionTd = document.createElement('td');
            actionTd.className = 'action-column-cell';
            const actionWrapper = document.createElement('div');
            actionWrapper.className = 'action-buttons-wrapper';
            
            const editBtn = document.createElement('button');
            editBtn.className = 'btn btn-secondary action-btn edit-row-btn';
            editBtn.textContent = '编辑';
            editBtn.dataset.row = JSON.stringify(row);
            
            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'btn btn-danger action-btn delete-row-btn';
            deleteBtn.textContent = '删除';
            deleteBtn.dataset.row = JSON.stringify(row);
            
            actionWrapper.appendChild(editBtn);
            actionWrapper.appendChild(deleteBtn);
            actionTd.appendChild(actionWrapper);
            bodyRow.appendChild(actionTd);
        }
        
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
function updatePagination(total, page, pageSize, isClickHouse = false) {
    if (isClickHouse) {
        // ClickHouse 不支持分页，只显示提示信息
        paginationInfo.textContent = `显示前 10 条数据（ClickHouse 不支持分页）`;
        pagination.innerHTML = '';
        return;
    }
    
    const totalPages = Math.ceil(total / pageSize);
    
    // 如果没有数据，显示提示并禁用所有分页按钮
    if (total === 0) {
        paginationInfo.textContent = '没有数据';
        pagination.innerHTML = `
            <button disabled>上一页</button>
            <button disabled>下一页</button>
        `;
        return;
    }
    
    paginationInfo.textContent = `共 ${total} 条，第 ${page}/${totalPages} 页`;
    
    let paginationHTML = '';
    // 上一页按钮：第一页或没有数据时禁用
    const prevDisabled = page === 1 || total === 0;
    paginationHTML += `<button ${prevDisabled ? 'disabled' : ''} onclick="changePage(${page - 1})">上一页</button>`;
    
    // 页码按钮
    for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
        paginationHTML += `<button class="${i === page ? 'active' : ''}" onclick="changePage(${i})">${i}</button>`;
    }
    
    // 下一页按钮：最后一页或没有数据时禁用
    const nextDisabled = page >= totalPages || total === 0;
    paginationHTML += `<button ${nextDisabled ? 'disabled' : ''} onclick="changePage(${page + 1})">下一页</button>`;
    pagination.innerHTML = paginationHTML;
}

// 切换页码
function changePage(page) {
    currentPage = page;
    loadTableData();
}

// 分页大小改变
pageSizeSelect.addEventListener('change', (e) => {
    const newPageSize = parseInt(e.target.value);
    pageSize = newPageSize;
    currentPage = 1; // 重置到第一页
    loadTableData();
});

// 刷新数据
refreshData.addEventListener('click', loadTableData);

// 加载表结构
async function loadTableSchema() {
    if (!currentTable) return;
    
    showLoading(schemaLoading);
    try {
        const response = await apiRequest(`${API_BASE}/table/schema?table=${currentTable}`);
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            showNotification(data.message || '加载表结构失败', 'error');
            hideLoading(schemaLoading);
            copySchemaBtn.style.display = 'none';
            return;
        }
        
        if (data.success) {
            schemaContent.textContent = data.schema;
            copySchemaBtn.style.display = 'block';
        }
    } catch (error) {
        showNotification('加载表结构失败: ' + error.message, 'error');
        copySchemaBtn.style.display = 'none';
    } finally {
        hideLoading(schemaLoading);
    }
}

// 复制表结构
copySchemaBtn.addEventListener('click', async () => {
    const schemaText = schemaContent.textContent;
    if (!schemaText || schemaText === '请选择一个表查看结构') {
        showNotification('没有可复制的内容', 'error');
        return;
    }
    
    try {
        await navigator.clipboard.writeText(schemaText);
        showNotification('表结构已复制到剪贴板', 'success');
    } catch (error) {
        // 降级方案：使用传统方法
        const textArea = document.createElement('textarea');
        textArea.value = schemaText;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            showNotification('表结构已复制到剪贴板', 'success');
        } catch (err) {
            showNotification('复制失败，请手动复制', 'error');
        }
        document.body.removeChild(textArea);
    }
});

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
    } else if (tabName === 'schema' && !currentTable) {
        // 如果没有选择表，隐藏复制按钮
        copySchemaBtn.style.display = 'none';
    }
}

// 执行SQL查询
executeQuery.addEventListener('click', async () => {
    const query = sqlQuery.value.trim();
    if (!query) {
        showNotification('请输入SQL查询', 'error');
        return;
    }
    
    showLoading(queryLoading);
    setButtonLoading(executeQuery, true);
    try {
        const response = await apiRequest(`${API_BASE}/query`, {
            method: 'POST',
            body: JSON.stringify({ query })
        });
        
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            queryResults.innerHTML = `<div class="query-message error">${data.message || '执行失败'}</div>`;
            return;
        }
        
        if (response.ok && data.success) {
            if (data.data) {
                // 查询结果
                displayQueryResults(data.data);
            } else if (data.affected !== undefined) {
                // 更新/删除/插入结果
                queryResults.innerHTML = `<div class="query-message success">操作成功，影响 ${data.affected} 行</div>`;
            }
        }
    } catch (error) {
        queryResults.innerHTML = `<div class="query-message error">执行失败: ${error.message}</div>`;
    } finally {
        hideLoading(queryLoading);
        setButtonLoading(executeQuery, false);
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
    apiRequest(`${API_BASE}/table/columns?table=${currentTable}`)
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
    const columns = await apiRequest(`${API_BASE}/table/columns?table=${currentTable}`)
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
        const response = await apiRequest(`${API_BASE}/row/update`, {
            method: 'POST',
            body: JSON.stringify({
                table: currentTable,
                data: updateData,
                where: where
            })
        });
        
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            showNotification(data.message || '更新失败', 'error');
            return;
        }
        
        if (response.ok && data.success) {
            showNotification('更新成功', 'success');
            editModal.style.display = 'none';
            loadTableData();
        }
    } catch (error) {
        showNotification('更新失败: ' + error.message, 'error');
    }
});

// 删除行（全局函数，供外部调用）
window.deleteRow = function(rowData) {
    currentRowData = rowData;
    
    // 获取主键列
    apiRequest(`${API_BASE}/table/columns?table=${currentTable}`)
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
    
    setButtonLoading(confirmDelete, true);
    try {
        const response = await apiRequest(`${API_BASE}/row/delete`, {
            method: 'POST',
            body: JSON.stringify({
                table: currentTable,
                where: currentDeleteWhere
            })
        });
        
        const data = await response.json();
        
        if (!response.ok || !data.success) {
            showNotification(data.message || '删除失败', 'error');
            return;
        }
        
        if (response.ok && data.success) {
            showNotification('删除成功', 'success');
            deleteModal.style.display = 'none';
            loadTableData();
        }
    } catch (error) {
        showNotification('删除失败: ' + error.message, 'error');
    } finally {
        setButtonLoading(confirmDelete, false);
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

