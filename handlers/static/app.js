// ==================== i18n 国际化支持 ====================
// 语言配置
const i18n = {
    currentLang: 'zh-CN', // 默认简体中文
    translations: {
        'en': {
            // 通用
            'common.loading': 'Loading...',
            'common.confirm': 'Confirm',
            'common.cancel': 'Cancel',
            'common.delete': 'Delete',
            'common.edit': 'Edit',
            'common.save': 'Save',
            'common.refresh': 'Refresh',
            'common.close': 'Close',
            'common.clear': 'Clear',
            'common.clearAll': 'Clear All',
            'common.switch': 'Switch',
            'common.disconnect': 'Disconnect',
            'common.connect': 'Connect',
            'common.connected': 'Connected',
            'common.disconnected': 'Disconnected',
            'common.noData': 'No Data',
            'common.operation': 'Operation',
            'common.null': 'NULL',
            
            // 连接管理
            'connection.management': 'Connection Management',
            'connection.new': '+ New Connection',
            'connection.newTitle': 'New Database Connection',
            'connection.noActive': 'No Active Connections',
            'connection.saved': 'Saved Connections',
            'connection.remember': 'Remember Connection',
            'connection.mode': 'Connection Mode',
            'connection.modeForm': 'Form Input',
            'connection.modeDSN': 'DSN Connection String',
            'connection.dbType': 'Database Type',
            'connection.host': 'Host',
            'connection.port': 'Port',
            'connection.user': 'Username',
            'connection.password': 'Password',
            'connection.database': 'Database',
            'connection.selectDatabase': 'Please select database...',
            'connection.success': 'Connection successful',
            'connection.failed': 'Connection failed',
            'connection.disconnected': 'Disconnected',
            'connection.switched': 'Switched to connection',
            'connection.notExists': 'Connection does not exist',
            'connection.noActiveConn': 'No active connection',
            'connection.id': 'Connection ID',
            'connection.noSaved': 'No saved connections',
            
            // 代理
            'proxy.use': 'Use Proxy (SSH, etc.)',
            'proxy.type': 'Proxy Type',
            'proxy.host': 'Proxy Host',
            'proxy.port': 'Proxy Port',
            'proxy.user': 'Proxy Username',
            'proxy.password': 'Proxy Password',
            'proxy.key': 'SSH Private Key (optional, base64 encoded)',
            'proxy.keyHint': 'If a private key is provided, key authentication will be prioritized',
            'proxy.required': 'Please fill in proxy host and username',
            
            // 数据库和表
            'db.select': 'Select Database',
            'db.tables': 'Data Tables',
            'db.noTables': 'No tables found',
            'db.filterTables': 'Filter table names...',
            'db.selectTable': 'Please select a table to view schema',
            
            // 数据标签页
            'tab.data': 'Data',
            'tab.schema': 'Schema',
            'tab.query': 'SQL Query',
            'data.perPage': 'Per Page:',
            'data.total': 'Total {total} records, Page {page}/{totalPages}',
            'data.clickhouseNoPagination': 'Showing first 10 records (ClickHouse does not support pagination)',
            'data.prevPage': 'Previous',
            'data.nextPage': 'Next',
            'data.copySchema': 'Copy',
            'data.copySchemaTitle': 'Copy Schema',
            
            // SQL查询
            'query.placeholder': 'Enter SQL query...',
            'query.execute': 'Execute Query',
            'query.empty': 'Please enter SQL query',
            'query.emptyResult': 'Query result is empty',
            'query.success': 'Operation successful, {affected} rows affected',
            'query.failed': 'Execution failed',
            'query.unsupported': 'Unsupported SQL type',
            'query.exportExcel': 'Export to Excel',
            'query.exportSuccess': 'Export successful',
            'query.history': 'Query History',
            'query.showHistory': 'History',
            'query.noHistory': 'No query history',
            'query.historyItem': 'History #{index}',
            'query.clearHistory': 'Clear History',
            'query.historyCleared': 'History cleared',
            
            // 编辑和删除
            'edit.title': 'Edit Row Data',
            'edit.save': 'Update successful',
            'edit.failed': 'Update failed',
            'delete.title': 'Confirm Delete',
            'delete.message': 'Are you sure you want to delete this row? This operation cannot be undone.',
            'delete.success': 'Delete successful',
            'delete.failed': 'Delete failed',
            'delete.connection': 'Confirm Delete Connection',
            'delete.connectionMessage': 'Are you sure you want to delete this saved connection? This operation cannot be undone.',
            'delete.connectionSuccess': 'Connection deleted',
            'delete.clearAll': 'Confirm Clear All Connections',
            'delete.clearAllMessage': 'Are you sure you want to clear all saved connections? This operation cannot be undone.',
            'delete.clearAllSuccess': 'All saved connections cleared',
            
            // 错误消息
            'error.selectDbType': 'Please select database type',
            'error.fillHostUser': 'Please fill in host and username',
            'error.enterDSN': 'Please enter DSN connection string',
            'error.loadDbTypes': 'Failed to load database types',
            'error.loadDatabases': 'Failed to load database list',
            'error.loadTables': 'Failed to load table list',
            'error.loadData': 'Failed to load data',
            'error.loadSchema': 'Failed to load schema',
            'error.loadColumns': 'Failed to load column information',
            'error.switchDatabase': 'Failed to switch database',
            'error.copyFailed': 'Copy failed, please copy manually',
            'error.copySuccess': 'Schema copied to clipboard',
            'error.noContent': 'No content to copy',
            'error.exportFailed': 'Export failed',
            'error.noTable': 'No table selected',
            
            // 语言切换
            'lang.en': 'English',
            'lang.zh-CN': '简体中文',
            'lang.zh-TW': '繁體中文',
            'lang.switch': 'Language'
        },
        'zh-CN': {
            // 通用
            'common.loading': '加载中...',
            'common.confirm': '确认',
            'common.cancel': '取消',
            'common.delete': '删除',
            'common.edit': '编辑',
            'common.save': '保存',
            'common.refresh': '刷新',
            'common.close': '关闭',
            'common.clear': '清空',
            'common.clearAll': '清除所有',
            'common.switch': '切换',
            'common.disconnect': '断开',
            'common.connect': '连接',
            'common.connected': '已连接',
            'common.disconnected': '未连接',
            'common.noData': '没有数据',
            'common.operation': '操作',
            'common.null': 'NULL',
            
            // 连接管理
            'connection.management': '连接管理',
            'connection.new': '+ 新增连接',
            'connection.newTitle': '新增数据库连接',
            'connection.noActive': '暂无活动连接',
            'connection.saved': '已保存的连接',
            'connection.remember': '记住连接',
            'connection.mode': '连接方式',
            'connection.modeForm': '表单输入',
            'connection.modeDSN': 'DSN连接字符串',
            'connection.dbType': '数据库类型',
            'connection.host': '主机',
            'connection.port': '端口',
            'connection.user': '用户名',
            'connection.password': '密码',
            'connection.database': '数据库',
            'connection.selectDatabase': '请选择数据库...',
            'connection.success': '连接成功',
            'connection.failed': '连接失败',
            'connection.disconnected': '已断开连接',
            'connection.switched': '已切换到连接',
            'connection.notExists': '连接不存在',
            'connection.noActiveConn': '没有活动连接',
            'connection.id': '连接ID',
            'connection.noSaved': '暂无保存的连接',
            
            // 代理
            'proxy.use': '使用代理（SSH等）',
            'proxy.type': '代理类型',
            'proxy.host': '代理主机',
            'proxy.port': '代理端口',
            'proxy.user': '代理用户名',
            'proxy.password': '代理密码',
            'proxy.key': 'SSH私钥（可选，base64编码）',
            'proxy.keyHint': '如果提供了私钥，将优先使用私钥认证',
            'proxy.required': '请填写代理主机和用户名',
            
            // 数据库和表
            'db.select': '选择数据库',
            'db.tables': '数据表',
            'db.noTables': '没有找到表',
            'db.filterTables': '筛选表名...',
            'db.selectTable': '请选择一个表查看结构',
            
            // 数据标签页
            'tab.data': '数据',
            'tab.schema': '结构',
            'tab.query': 'SQL查询',
            'data.perPage': '每页:',
            'data.total': '共 {total} 条，第 {page}/{totalPages} 页',
            'data.clickhouseNoPagination': '显示前 10 条数据（ClickHouse 不支持分页）',
            'data.prevPage': '上一页',
            'data.nextPage': '下一页',
            'data.copySchema': '复制',
            'data.copySchemaTitle': '复制结构',
            'data.exportExcel': '导出Excel',
            'data.exportSuccess': '导出成功',
            
            // SQL查询
            'query.placeholder': '输入SQL查询...',
            'query.execute': '执行查询',
            'query.empty': '请输入SQL查询',
            'query.emptyResult': '查询结果为空',
            'query.success': '操作成功，影响 {affected} 行',
            'query.failed': '执行失败',
            'query.unsupported': '不支持的SQL类型',
            'query.exportExcel': '导出Excel',
            'query.exportSuccess': '导出成功',
            'query.history': '查询历史',
            'query.showHistory': '历史',
            'query.noHistory': '暂无查询历史',
            'query.historyItem': '历史 #{index}',
            'query.clearHistory': '清空历史',
            'query.historyCleared': '历史记录已清空',
            'query.format': '格式化',
            'query.formatSuccess': 'SQL格式化成功',
            'query.formatFailed': '格式化失败',
            'query.formatterNotLoaded': 'SQL格式化库未加载',
            
            // 编辑和删除
            'edit.title': '编辑行数据',
            'edit.save': '更新成功',
            'edit.failed': '更新失败',
            'delete.title': '确认删除',
            'delete.message': '确定要删除这行数据吗？此操作无法撤销。',
            'delete.success': '删除成功',
            'delete.failed': '删除失败',
            'delete.connection': '确认删除连接',
            'delete.connectionMessage': '确定要删除这个保存的连接吗？此操作无法撤销。',
            'delete.connectionSuccess': '已删除连接',
            'delete.clearAll': '确认清除所有连接',
            'delete.clearAllMessage': '确定要清除所有保存的连接吗？此操作无法撤销。',
            'delete.clearAllSuccess': '已清空所有保存的连接',
            
            // 错误消息
            'error.selectDbType': '请选择数据库类型',
            'error.fillHostUser': '请填写主机和用户名',
            'error.enterDSN': '请输入DSN连接字符串',
            'error.loadDbTypes': '加载数据库类型失败',
            'error.loadDatabases': '获取数据库列表失败',
            'error.loadTables': '加载表列表失败',
            'error.loadData': '获取数据失败',
            'error.loadSchema': '加载表结构失败',
            'error.loadColumns': '获取列信息失败',
            'error.switchDatabase': '切换数据库失败',
            'error.copyFailed': '复制失败，请手动复制',
            'error.copySuccess': '表结构已复制到剪贴板',
            'error.noContent': '没有可复制的内容',
            'error.exportFailed': '导出失败',
            'error.noTable': '未选择表',
            
            // 语言切换
            'lang.en': 'English',
            'lang.zh-CN': '简体中文',
            'lang.zh-TW': '繁體中文',
            'lang.switch': '语言'
        },
        'zh-TW': {
            // 通用
            'common.loading': '載入中...',
            'common.confirm': '確認',
            'common.cancel': '取消',
            'common.delete': '刪除',
            'common.edit': '編輯',
            'common.save': '儲存',
            'common.refresh': '重新整理',
            'common.close': '關閉',
            'common.clear': '清空',
            'common.clearAll': '清除所有',
            'common.switch': '切換',
            'common.disconnect': '斷開',
            'common.connect': '連接',
            'common.connected': '已連接',
            'common.disconnected': '未連接',
            'common.noData': '沒有資料',
            'common.operation': '操作',
            'common.null': 'NULL',
            
            // 连接管理
            'connection.management': '連接管理',
            'connection.new': '+ 新增連接',
            'connection.newTitle': '新增資料庫連接',
            'connection.noActive': '暫無活動連接',
            'connection.saved': '已儲存的連接',
            'connection.remember': '記住連接',
            'connection.mode': '連接方式',
            'connection.modeForm': '表單輸入',
            'connection.modeDSN': 'DSN連接字串',
            'connection.dbType': '資料庫類型',
            'connection.host': '主機',
            'connection.port': '埠號',
            'connection.user': '使用者名稱',
            'connection.password': '密碼',
            'connection.database': '資料庫',
            'connection.selectDatabase': '請選擇資料庫...',
            'connection.success': '連接成功',
            'connection.failed': '連接失敗',
            'connection.disconnected': '已斷開連接',
            'connection.switched': '已切換到連接',
            'connection.notExists': '連接不存在',
            'connection.noActiveConn': '沒有活動連接',
            'connection.id': '連接ID',
            'connection.noSaved': '暫無儲存的連接',
            
            // 代理
            'proxy.use': '使用代理（SSH等）',
            'proxy.type': '代理類型',
            'proxy.host': '代理主機',
            'proxy.port': '代理埠號',
            'proxy.user': '代理使用者名稱',
            'proxy.password': '代理密碼',
            'proxy.key': 'SSH私鑰（可選，base64編碼）',
            'proxy.keyHint': '如果提供了私鑰，將優先使用私鑰認證',
            'proxy.required': '請填寫代理主機和使用者名稱',
            
            // 数据库和表
            'db.select': '選擇資料庫',
            'db.tables': '資料表',
            'db.noTables': '沒有找到表',
            'db.filterTables': '篩選表名...',
            'db.selectTable': '請選擇一個表查看結構',
            
            // 数据标签页
            'tab.data': '資料',
            'tab.schema': '結構',
            'tab.query': 'SQL查詢',
            'data.perPage': '每頁:',
            'data.total': '共 {total} 筆，第 {page}/{totalPages} 頁',
            'data.clickhouseNoPagination': '顯示前 10 筆資料（ClickHouse 不支援分頁）',
            'data.prevPage': '上一頁',
            'data.nextPage': '下一頁',
            'data.copySchema': '複製',
            'data.copySchemaTitle': '複製結構',
            'data.exportExcel': '匯出Excel',
            'data.exportSuccess': '匯出成功',
            
            // SQL查询
            'query.placeholder': '輸入SQL查詢...',
            'query.execute': '執行查詢',
            'query.empty': '請輸入SQL查詢',
            'query.emptyResult': '查詢結果為空',
            'query.success': '操作成功，影響 {affected} 行',
            'query.failed': '執行失敗',
            'query.unsupported': '不支援的SQL類型',
            'query.exportExcel': '匯出Excel',
            'query.exportSuccess': '匯出成功',
            'query.history': '查詢歷史',
            'query.showHistory': '歷史',
            'query.noHistory': '暫無查詢歷史',
            'query.historyItem': '歷史 #{index}',
            'query.clearHistory': '清空歷史',
            'query.historyCleared': '歷史記錄已清空',
            'query.format': '格式化',
            'query.formatSuccess': 'SQL格式化成功',
            'query.formatFailed': '格式化失敗',
            'query.formatterNotLoaded': 'SQL格式化庫未載入',
            
            // 编辑和删除
            'edit.title': '編輯行資料',
            'edit.save': '更新成功',
            'edit.failed': '更新失敗',
            'delete.title': '確認刪除',
            'delete.message': '確定要刪除這行資料嗎？此操作無法復原。',
            'delete.success': '刪除成功',
            'delete.failed': '刪除失敗',
            'delete.connection': '確認刪除連接',
            'delete.connectionMessage': '確定要刪除這個儲存的連接嗎？此操作無法復原。',
            'delete.connectionSuccess': '已刪除連接',
            'delete.clearAll': '確認清除所有連接',
            'delete.clearAllMessage': '確定要清除所有儲存的連接嗎？此操作無法復原。',
            'delete.clearAllSuccess': '已清空所有儲存的連接',
            
            // 错误消息
            'error.selectDbType': '請選擇資料庫類型',
            'error.fillHostUser': '請填寫主機和使用者名稱',
            'error.enterDSN': '請輸入DSN連接字串',
            'error.loadDbTypes': '載入資料庫類型失敗',
            'error.loadDatabases': '取得資料庫列表失敗',
            'error.loadTables': '載入表列表失敗',
            'error.loadData': '取得資料失敗',
            'error.loadSchema': '載入表結構失敗',
            'error.loadColumns': '取得欄位資訊失敗',
            'error.switchDatabase': '切換資料庫失敗',
            'error.copyFailed': '複製失敗，請手動複製',
            'error.copySuccess': '表結構已複製到剪貼簿',
            'error.noContent': '沒有可複製的內容',
            'error.exportFailed': '匯出失敗',
            'error.noTable': '未選擇表',
            
            // 语言切换
            'lang.en': 'English',
            'lang.zh-CN': '简体中文',
            'lang.zh-TW': '繁體中文',
            'lang.switch': '語言'
        }
    },
    
    // 翻译函数
    t(key, params = {}) {
        const lang = this.currentLang;
        const translation = this.translations[lang]?.[key] || key;
        
        // 支持参数替换 {param}
        return translation.replace(/\{(\w+)\}/g, (match, param) => {
            return params[param] !== undefined ? params[param] : match;
        });
    },
    
    // 设置语言
    setLanguage(lang) {
        if (this.translations[lang]) {
            this.currentLang = lang;
            localStorage.setItem('simple-db-web-lang', lang);
            document.documentElement.lang = lang === 'en' ? 'en' : (lang === 'zh-TW' ? 'zh-TW' : 'zh-CN');
            this.updateUI();
        }
    },
    
    // 初始化语言
    init() {
        const savedLang = localStorage.getItem('simple-db-web-lang');
        if (savedLang && this.translations[savedLang]) {
            // 如果 localStorage 中有保存的语言，使用保存的语言
            this.currentLang = savedLang;
        } else {
            // 默认使用简体中文
            this.currentLang = 'zh-CN';
            // 保存默认语言到 localStorage
            localStorage.setItem('simple-db-web-lang', 'zh-CN');
        }
        document.documentElement.lang = this.currentLang === 'en' ? 'en' : (this.currentLang === 'zh-TW' ? 'zh-TW' : 'zh-CN');
    },
    
    // 更新UI文本
    updateUI() {
        // 触发自定义事件，让其他代码更新文本
        window.dispatchEvent(new CustomEvent('languageChanged', { detail: { lang: this.currentLang } }));
    }
};

// 简化的翻译函数
function t(key, params = {}) {
    return i18n.t(key, params);
}

// 导出到全局
window.i18n = i18n;
window.t = t;

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
const exportDataBtn = document.getElementById('exportDataBtn');
const pagination = document.getElementById('pagination');
const paginationInfo = document.getElementById('paginationInfo');
const pageSizeSelect = document.getElementById('pageSizeSelect');
const schemaContent = document.getElementById('schemaContent');
const copySchemaBtn = document.getElementById('copySchemaBtn');
const sqlQuery = document.getElementById('sqlQuery');
const executeQuery = document.getElementById('executeQuery');
const clearQuery = document.getElementById('clearQuery');
const exportQueryBtn = document.getElementById('exportQueryBtn');
const showHistoryBtn = document.getElementById('showHistoryBtn');
const formatQueryBtn = document.getElementById('formatQueryBtn');
const queryHistoryModal = document.getElementById('queryHistoryModal');
const queryHistoryList = document.getElementById('queryHistoryList');
const closeQueryHistoryModal = document.getElementById('closeQueryHistoryModal');
const closeQueryHistoryBtn = document.getElementById('closeQueryHistoryBtn');
const clearQueryHistory = document.getElementById('clearQueryHistory');

// CodeMirror编辑器实例
let sqlEditor = null;
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

// 语言切换相关
const languageSelect = document.getElementById('languageSelect');

// 更新所有带有 data-i18n 属性的元素
function updateI18nElements() {
    // 更新 textContent
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (key && !el.hasAttribute('data-i18n-ignore')) {
            el.textContent = t(key);
        }
    });
    
    // 更新 placeholder
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (key) {
            el.placeholder = t(key);
        }
    });
    
    // 更新 title
    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        const key = el.getAttribute('data-i18n-title');
        if (key) {
            el.title = t(key);
        }
    });
    
    // 更新 value（用于 select option）
    document.querySelectorAll('[data-i18n-value]').forEach(el => {
        const key = el.getAttribute('data-i18n-value');
        if (key) {
            el.value = t(key);
        }
    });
}

// 语言切换事件
if (languageSelect) {
    languageSelect.addEventListener('change', (e) => {
        i18n.setLanguage(e.target.value);
        updateI18nElements();
        // 更新语言选择器的值
        languageSelect.value = i18n.currentLang;
    });
}

// 监听语言变化事件
window.addEventListener('languageChanged', () => {
    updateI18nElements();
    if (languageSelect) {
        languageSelect.value = i18n.currentLang;
    }
});

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
        deleteBtn.title = t('common.delete');
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
            const newConnectionId = data.connectionId;
            const connInfo = {
                type: savedConn.type || 'mysql',
                host: savedConn.host || '',
                port: savedConn.port || '3306',
                user: savedConn.user || '',
                dsn: savedConn.dsn || '',
                proxy: savedConn.proxy || null
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
            currentDbType = savedConn.type || 'mysql'; // 保存数据库类型
            
            // 保存到sessionStorage（用于页面刷新后恢复）
            sessionStorage.setItem('currentConnectionId', newConnectionId);
            sessionStorage.setItem('currentConnectionInfo', JSON.stringify(connInfo));
            
            // 更新UI
            updateConnectionStatus(true);
            updateConnectionInfo(connInfo);
            updateActiveConnectionsList();
            
            // 检查DSN中是否包含数据库
            const dsn = connInfo.dsn || '';
            const hasDatabaseInDSN = dsn && (dsn.includes('/') && !dsn.endsWith('/') && !dsn.endsWith('/?'));
            
            if (hasDatabaseInDSN) {
                // DSN中包含数据库,直接使用该数据库
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
                databasePanel.style.display = 'block';
                await loadDatabases(data.databases || []);
            }
            showNotification(t('connection.success'), 'success');
        } else {
            showNotification(data.message || t('connection.failed'), 'error');
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
        showNotification(t('delete.connectionSuccess'), 'success');
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
    showNotification(t('delete.clearAllSuccess'), 'success');
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

// 初始化CodeMirror编辑器
function initCodeMirror() {
    if (typeof CodeMirror === 'undefined') {
        console.warn('CodeMirror未加载，使用普通textarea');
        return;
    }
    
    // 获取数据库表和列信息用于自动补全
    let tables = {};
    if (allTables && allTables.length > 0) {
        allTables.forEach(table => {
            tables[table] = currentColumns || [];
        });
    }
    
    // 隐藏原始textarea
    sqlQuery.style.display = 'none';
    
    sqlEditor = CodeMirror.fromTextArea(sqlQuery, {
        mode: 'text/x-sql',
        theme: 'monokai',
        lineNumbers: true,
        lineWrapping: true,
        indentWithTabs: true,
        smartIndent: true,
        matchBrackets: true,
        autofocus: false,
        extraKeys: {
            'Ctrl-Space': function(cm) {
                CodeMirror.commands.autocomplete(cm);
            },
            'Tab': function(cm) {
                if (cm.somethingSelected()) {
                    cm.indentSelection('add');
                } else {
                    cm.replaceSelection('  ', 'end');
                }
            }
        },
        hintOptions: {
            tables: tables,
            completeSingle: false,
            completeOnSingleClick: false
        }
    });
    
    // 设置编辑器样式和容器
    const container = document.getElementById('sqlEditorContainer');
    if (container) {
        // 确保CodeMirror在容器内正确显示
        sqlEditor.setSize('100%', '150px');
    }
    
    // 更新自动补全表信息的函数
    function updateHintTables() {
        if (allTables && allTables.length > 0) {
            let tablesForHint = {};
            allTables.forEach(table => {
                tablesForHint[table] = currentColumns || [];
            });
            sqlEditor.setOption('hintOptions', {
                tables: tablesForHint,
                completeSingle: false,
                completeOnSingleClick: false
            });
        }
    }
    
    // 监听编辑器内容变化，更新自动补全的表信息
    sqlEditor.on('focus', updateHintTables);
    
    // 监听输入，自动触发补全提示
    sqlEditor.on('inputRead', function(cm) {
        // 延迟触发自动补全，避免过于频繁
        clearTimeout(cm.state.completionTimeout);
        cm.state.completionTimeout = setTimeout(function() {
            if (!cm.state.completionActive) {
                // 更新表信息
                updateHintTables();
                // 触发自动补全
                CodeMirror.commands.autocomplete(cm, null, {completeSingle: false});
            }
        }, 300);
    });
}

// 查询历史记录管理
const queryHistory = {
    // 保存查询历史（最多10条）
    save(query) {
        if (!query || !query.trim()) return;
        
        let history = this.load();
        // 移除重复项
        history = history.filter(item => item !== query.trim());
        // 添加到开头
        history.unshift(query.trim());
        // 只保留最近10条
        if (history.length > 10) {
            history = history.slice(0, 10);
        }
        localStorage.setItem('sqlQueryHistory', JSON.stringify(history));
    },
    
    // 加载查询历史
    load() {
        try {
            const history = localStorage.getItem('sqlQueryHistory');
            return history ? JSON.parse(history) : [];
        } catch (e) {
            console.error('加载查询历史失败:', e);
            return [];
        }
    },
    
    // 清空查询历史
    clear() {
        localStorage.removeItem('sqlQueryHistory');
    },
    
    // 显示查询历史
    display() {
        const history = this.load();
        queryHistoryList.innerHTML = '';
        
        if (history.length === 0) {
            queryHistoryList.innerHTML = `<div style="padding: 2rem; color: var(--text-secondary); text-align: center; font-size: 0.875rem;" data-i18n="query.noHistory">暂无查询历史</div>`;
            updateI18nElements();
            return;
        }
        
        history.forEach((query, index) => {
            const item = document.createElement('div');
            item.className = 'query-history-item';
            item.style.cssText = 'padding: 1rem; border-bottom: 1px solid var(--border-color); cursor: pointer; transition: background 0.2s;';
            item.innerHTML = `
                <div style="font-size: 0.875rem; color: var(--text-primary); margin-bottom: 0.5rem; word-break: break-all; white-space: pre-wrap; font-family: monospace; background: var(--surface-light); padding: 0.75rem; border-radius: 4px;">${escapeHtml(query)}</div>
                <div style="font-size: 0.75rem; color: var(--text-secondary);">${t('query.historyItem', { index: index + 1 })}</div>
            `;
            
            item.addEventListener('click', () => {
                if (sqlEditor) {
                    sqlEditor.setValue(query);
                    sqlEditor.focus();
                } else {
                    sqlQuery.value = query;
                }
                if (queryHistoryModal) {
                    queryHistoryModal.style.display = 'none';
                }
            });
            
            item.addEventListener('mouseenter', () => {
                item.style.background = 'var(--surface-light)';
            });
            
            item.addEventListener('mouseleave', () => {
                item.style.background = 'transparent';
            });
            
            queryHistoryList.appendChild(item);
        });
    }
};

// 页面加载完成后初始化 i18n 和恢复连接
document.addEventListener('DOMContentLoaded', () => {
    // 初始化 i18n（从 localStorage 读取或使用默认值）
    i18n.init();
    
    // 更新所有翻译元素
    updateI18nElements();
    
    // 确保语言选择框的值正确设置
    const langSelect = document.getElementById('languageSelect');
    if (langSelect) {
        langSelect.value = i18n.currentLang;
    }
    
    // 初始化CodeMirror编辑器
    initCodeMirror();
    
    // 恢复连接
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
            showNotification(t('error.enterDSN'), 'error');
            return;
        }
    } else {
        const hostInput = document.getElementById('host');
        const userInput = document.getElementById('user');
        if (!hostInput || !hostInput.value || !userInput || !userInput.value) {
            showNotification(t('error.fillHostUser'), 'error');
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
            showNotification(t('proxy.required'), 'error');
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
                    <div style="font-size: 0.75rem; color: var(--text-secondary);">${t('connection.id')}: ${connId.substring(0, 8)}...</div>
                </div>
                <div style="display: flex; gap: 0.5rem;">
                    <button class="btn btn-secondary switch-connection-btn" data-connection-id="${connId}" style="font-size: 0.75rem; padding: 0.25rem 0.5rem;">${t('common.switch')}</button>
                    <button class="btn btn-danger disconnect-connection-btn" data-connection-id="${connId}" style="font-size: 0.75rem; padding: 0.25rem 0.5rem;">${t('common.disconnect')}</button>
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
        showNotification(t('connection.notExists'), 'error');
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
    
    showNotification(t('connection.switched'), 'success');
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
                databaseSelect.innerHTML = `<option value="">${t('connection.selectDatabase')}</option>`;
                tableFilter.value = '';
                allTables = [];
                currentColumns = [];
            }
            
            updateActiveConnectionsList();
            showNotification(t('connection.disconnected'), 'success');
        } else {
            showNotification(data.message || t('connection.failed'), 'error');
        }
    } catch (error) {
        showNotification(t('connection.failed') + ': ' + error.message, 'error');
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
        text.setAttribute('data-i18n', 'common.connected');
        text.textContent = t('common.connected');
    } else {
        indicator.classList.remove('connected');
        indicator.classList.add('disconnected');
        text.setAttribute('data-i18n', 'common.disconnected');
        text.textContent = t('common.disconnected');
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
            showNotification(t('error.loadDatabases') + ': ' + error.message, 'error');
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
            showNotification(t('connection.switched'), 'success');
            // 加载表列表
            if (data.tables) {
                displayTables(data.tables);
            } else {
                await loadTables();
            }
        } else {
            showNotification(data.message || t('error.switchDatabase'), 'error');
        }
    } catch (error) {
        showNotification(t('error.switchDatabase') + ': ' + error.message, 'error');
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
    
    // 更新CodeMirror编辑器的自动补全表信息
    if (sqlEditor && allTables && allTables.length > 0) {
        let tablesForHint = {};
        allTables.forEach(table => {
            tablesForHint[table] = currentColumns || [];
        });
        sqlEditor.setOption('hintOptions', {
            tables: tablesForHint,
            completeSingle: false
        });
    }
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
            showNotification(data.message || t('error.loadTables'), 'error');
            hideLoading(tablesLoading);
            setButtonLoading(refreshTables, false);
            return;
        }
        
        if (data.success) {
            displayTables(data.tables || []);
        }
    } catch (error) {
        showNotification(t('error.loadTables') + ': ' + error.message, 'error');
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
            showNotification(columnsData.message || t('error.loadColumns'), 'error');
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
            showNotification(data.message || t('error.loadData'), 'error');
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
            
            // 显示导出按钮
            if (exportDataBtn) {
                exportDataBtn.style.display = 'inline-block';
            }
        }
    } catch (error) {
        showNotification(t('error.loadData') + ': ' + error.message, 'error');
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
        emptyCell.textContent = t('common.noData');
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
            nullSpan.textContent = t('common.null');
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
            editBtn.textContent = t('common.edit');
            editBtn.dataset.row = JSON.stringify(row);
            
            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'btn btn-danger action-btn delete-row-btn';
            deleteBtn.textContent = t('common.delete');
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
        paginationInfo.textContent = t('data.clickhouseNoPagination');
        pagination.innerHTML = '';
        return;
    }
    
    const totalPages = Math.ceil(total / pageSize);
    
    // 如果没有数据，显示提示并禁用所有分页按钮
    if (total === 0) {
        paginationInfo.textContent = t('common.noData');
        pagination.innerHTML = `
            <button disabled>${t('data.prevPage')}</button>
            <button disabled>${t('data.nextPage')}</button>
        `;
        return;
    }
    
    paginationInfo.textContent = t('data.total', { total, page, totalPages });
    
    let paginationHTML = '';
    // 上一页按钮：第一页或没有数据时禁用
    const prevDisabled = page === 1 || total === 0;
    paginationHTML += `<button ${prevDisabled ? 'disabled' : ''} onclick="changePage(${page - 1})">${t('data.prevPage')}</button>`;
    
    // 页码按钮
    for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
        paginationHTML += `<button class="${i === page ? 'active' : ''}" onclick="changePage(${i})">${i}</button>`;
    }
    
    // 下一页按钮：最后一页或没有数据时禁用
    const nextDisabled = page >= totalPages || total === 0;
    paginationHTML += `<button ${nextDisabled ? 'disabled' : ''} onclick="changePage(${page + 1})">${t('data.nextPage')}</button>`;
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
            showNotification(data.message || t('error.loadSchema'), 'error');
            hideLoading(schemaLoading);
            copySchemaBtn.style.display = 'none';
            return;
        }
        
        if (data.success) {
            schemaContent.textContent = data.schema;
            copySchemaBtn.style.display = 'block';
            copySchemaBtn.setAttribute('data-i18n', 'data.copySchema');
            copySchemaBtn.setAttribute('data-i18n-title', 'data.copySchemaTitle');
            copySchemaBtn.textContent = t('data.copySchema');
            copySchemaBtn.title = t('data.copySchemaTitle');
        }
    } catch (error) {
        showNotification(t('error.loadSchema') + ': ' + error.message, 'error');
        copySchemaBtn.style.display = 'none';
    } finally {
        hideLoading(schemaLoading);
    }
}

// 复制表结构
copySchemaBtn.addEventListener('click', async () => {
    const schemaText = schemaContent.textContent;
    const selectTableText = t('db.selectTable');
    if (!schemaText || schemaText === selectTableText) {
        showNotification(t('error.noContent'), 'error');
        return;
    }
    
    try {
        await navigator.clipboard.writeText(schemaText);
        showNotification(t('error.copySuccess'), 'success');
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
            showNotification(t('error.copySuccess'), 'success');
        } catch (err) {
            showNotification(t('error.copyFailed'), 'error');
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
    const query = sqlEditor ? sqlEditor.getValue().trim() : sqlQuery.value.trim();
    if (!query) {
            showNotification(t('query.empty'), 'error');
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
            queryResults.innerHTML = `<div class="query-message error">${data.message || t('query.failed')}</div>`;
            // 隐藏导出按钮（查询失败）
            if (exportQueryBtn) {
                exportQueryBtn.style.display = 'none';
            }
            return;
        }
        
        if (response.ok && data.success) {
            // 保存查询历史
            queryHistory.save(query);
            
            if (data.data) {
                // 查询结果
                displayQueryResults(data.data);
                // 显示导出按钮
                if (exportQueryBtn) {
                    exportQueryBtn.style.display = 'inline-block';
                }
            } else if (data.affected !== undefined) {
                // 更新/删除/插入结果
                queryResults.innerHTML = `<div class="query-message success">${t('query.success', { affected: data.affected })}</div>`;
                // 隐藏导出按钮（非SELECT查询）
                if (exportQueryBtn) {
                    exportQueryBtn.style.display = 'none';
                }
            }
        }
    } catch (error) {
        queryResults.innerHTML = `<div class="query-message error">${t('query.failed')}: ${error.message}</div>`;
        // 隐藏导出按钮（查询失败）
        if (exportQueryBtn) {
            exportQueryBtn.style.display = 'none';
        }
    } finally {
        hideLoading(queryLoading);
        setButtonLoading(executeQuery, false);
    }
});

// 显示查询结果
function displayQueryResults(rows) {
    if (rows.length === 0) {
        queryResults.innerHTML = `<div class="query-message">${t('query.emptyResult')}</div>`;
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

// 格式化SQL查询
if (formatQueryBtn) {
    formatQueryBtn.addEventListener('click', () => {
        try {
            const query = sqlEditor ? sqlEditor.getValue() : sqlQuery.value;
            if (!query || !query.trim()) {
                showNotification(t('query.empty'), 'error');
                return;
            }
            
            // 尝试使用sql-formatter库
            let formatted;
            if (typeof sqlFormatter !== 'undefined' && sqlFormatter.format) {
                formatted = sqlFormatter.format(query, {
                    language: 'sql',
                    indent: '  ',
                    uppercase: false,
                    linesBetweenQueries: 2
                });
            } else if (typeof window.sqlFormatter !== 'undefined' && window.sqlFormatter.format) {
                formatted = window.sqlFormatter.format(query, {
                    language: 'sql',
                    indent: '  ',
                    uppercase: false,
                    linesBetweenQueries: 2
                });
            } else {
                // 如果库未加载，使用简单的格式化
                formatted = query
                    .replace(/\s+/g, ' ')
                    .replace(/\s*,\s*/g, ', ')
                    .replace(/\s*\(\s*/g, ' (')
                    .replace(/\s*\)\s*/g, ') ')
                    .replace(/\s*=\s*/g, ' = ')
                    .replace(/\s*SELECT\s+/gi, '\nSELECT ')
                    .replace(/\s*FROM\s+/gi, '\nFROM ')
                    .replace(/\s*WHERE\s+/gi, '\nWHERE ')
                    .replace(/\s*JOIN\s+/gi, '\nJOIN ')
                    .replace(/\s*LEFT\s+JOIN\s+/gi, '\nLEFT JOIN ')
                    .replace(/\s*RIGHT\s+JOIN\s+/gi, '\nRIGHT JOIN ')
                    .replace(/\s*INNER\s+JOIN\s+/gi, '\nINNER JOIN ')
                    .replace(/\s*GROUP\s+BY\s+/gi, '\nGROUP BY ')
                    .replace(/\s*ORDER\s+BY\s+/gi, '\nORDER BY ')
                    .replace(/\s*HAVING\s+/gi, '\nHAVING ')
                    .replace(/\s*LIMIT\s+/gi, '\nLIMIT ')
                    .trim();
            }
            
            if (sqlEditor) {
                sqlEditor.setValue(formatted);
                sqlEditor.focus();
            } else {
                sqlQuery.value = formatted;
            }
            
            showNotification(t('query.formatSuccess'), 'success');
        } catch (error) {
            console.error('格式化SQL失败:', error);
            showNotification(t('query.formatFailed') + ': ' + error.message, 'error');
        }
    });
}

// 清空查询
clearQuery.addEventListener('click', () => {
    if (sqlEditor) {
        sqlEditor.setValue('');
        sqlEditor.focus();
    } else {
    sqlQuery.value = '';
    }
    queryResults.innerHTML = '';
    // 隐藏导出按钮
    if (exportQueryBtn) {
        exportQueryBtn.style.display = 'none';
    }
});

// 显示/隐藏查询历史模态框
if (showHistoryBtn) {
    showHistoryBtn.addEventListener('click', () => {
        queryHistory.display();
        if (queryHistoryModal) {
            queryHistoryModal.style.display = 'flex';
        }
    });
}

// 关闭查询历史模态框
if (closeQueryHistoryModal) {
    closeQueryHistoryModal.addEventListener('click', () => {
        if (queryHistoryModal) {
            queryHistoryModal.style.display = 'none';
        }
    });
}

if (closeQueryHistoryBtn) {
    closeQueryHistoryBtn.addEventListener('click', () => {
        if (queryHistoryModal) {
            queryHistoryModal.style.display = 'none';
        }
    });
}

// 清空查询历史
if (clearQueryHistory) {
    clearQueryHistory.addEventListener('click', () => {
        queryHistory.clear();
        queryHistory.display();
        showNotification(t('query.historyCleared'), 'success');
    });
}

// 导出表数据为Excel
if (exportDataBtn) {
    exportDataBtn.addEventListener('click', async () => {
        if (!currentTable) {
            showNotification(t('error.noTable'), 'error');
            return;
        }
        
        setButtonLoading(exportDataBtn, true);
        try {
            const url = `${API_BASE}/table/export?table=${encodeURIComponent(currentTable)}&page=${currentPage}&pageSize=${pageSize}`;
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'X-Connection-ID': connectionId || ''
                }
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                showNotification(errorData.message || t('error.exportFailed'), 'error');
                return;
            }
            
            // 获取文件名
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = `${currentTable}_page${currentPage}_${new Date().toISOString().slice(0, 10)}.xlsx`;
            if (contentDisposition) {
                const filenameMatch = contentDisposition.match(/filename=(.+)/);
                if (filenameMatch) {
                    filename = filenameMatch[1];
                }
            }
            
            // 下载文件
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(downloadUrl);
            
            showNotification(t('data.exportSuccess'), 'success');
        } catch (error) {
            showNotification(t('error.exportFailed') + ': ' + error.message, 'error');
        } finally {
            setButtonLoading(exportDataBtn, false);
        }
    });
}

// 导出查询结果为Excel
if (exportQueryBtn) {
    exportQueryBtn.addEventListener('click', async () => {
        const query = sqlEditor ? sqlEditor.getValue().trim() : sqlQuery.value.trim();
        if (!query) {
            showNotification(t('query.empty'), 'error');
            return;
        }
        
        setButtonLoading(exportQueryBtn, true);
        try {
            const response = await fetch(`${API_BASE}/query/export`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Connection-ID': connectionId || ''
                },
                body: JSON.stringify({ query })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                showNotification(errorData.message || t('error.exportFailed'), 'error');
                return;
            }
            
            // 获取文件名
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = `query_result_${new Date().toISOString().slice(0, 10)}.xlsx`;
            if (contentDisposition) {
                const filenameMatch = contentDisposition.match(/filename=(.+)/);
                if (filenameMatch) {
                    filename = filenameMatch[1];
                }
            }
            
            // 下载文件
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(downloadUrl);
            
            showNotification(t('query.exportSuccess'), 'success');
        } catch (error) {
            showNotification(t('error.exportFailed') + ': ' + error.message, 'error');
        } finally {
            setButtonLoading(exportQueryBtn, false);
        }
    });
}

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
            showNotification(data.message || t('edit.failed'), 'error');
            return;
        }
        
        if (response.ok && data.success) {
            showNotification(t('edit.save'), 'success');
            editModal.style.display = 'none';
            loadTableData();
        }
    } catch (error) {
        showNotification(t('edit.failed') + ': ' + error.message, 'error');
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
            showNotification(data.message || t('delete.failed'), 'error');
            return;
        }
        
        if (response.ok && data.success) {
            showNotification(t('delete.success'), 'success');
            deleteModal.style.display = 'none';
            loadTableData();
        }
    } catch (error) {
        showNotification(t('delete.failed') + ': ' + error.message, 'error');
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


