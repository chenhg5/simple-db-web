package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/chenhg5/simple-db-web/database"
	"golang.org/x/crypto/ssh"
)

// SSHProxyConfig SSH代理配置
type SSHProxyConfig struct {
	Host     string `json:"host"`     // SSH服务器地址
	Port     string `json:"port"`     // SSH服务器端口，默认22
	User     string `json:"user"`     // SSH用户名
	Password string `json:"password"` // SSH密码（如果使用密码认证）
	KeyFile  string `json:"key_file"` // SSH私钥文件路径（如果使用密钥认证）
	KeyData  string `json:"key_data"` // SSH私钥内容（base64编码，如果使用密钥认证）
}

// SSHProxy SSH代理实现
type SSHProxy struct {
	client *ssh.Client
	config *SSHProxyConfig
}

// NewSSHProxy 创建SSH代理
// config: 代理配置的JSON字符串
func NewSSHProxy(config string) (Proxy, error) {
	var proxyConfig SSHProxyConfig
	if err := json.Unmarshal([]byte(config), &proxyConfig); err != nil {
		return nil, fmt.Errorf("解析SSH代理配置失败: %w", err)
	}

	// 设置默认端口
	if proxyConfig.Port == "" {
		proxyConfig.Port = "22"
	}

	// 构建SSH客户端配置
	sshConfig := &ssh.ClientConfig{
		User:            proxyConfig.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         10 * time.Second,
	}

	// 认证方式：优先使用密钥，其次使用密码
	if proxyConfig.KeyData != "" {
		// 使用提供的密钥数据
		signer, err := ssh.ParsePrivateKey([]byte(proxyConfig.KeyData))
		if err != nil {
			return nil, fmt.Errorf("解析SSH私钥失败: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if proxyConfig.KeyFile != "" {
		// 从文件读取密钥
		// 注意：在生产环境中，应该从安全的位置读取密钥文件
		// 这里简化处理，实际使用时应该由外部项目提供密钥内容
		return nil, fmt.Errorf("从文件读取SSH密钥暂不支持，请使用key_data字段提供密钥内容")
	} else if proxyConfig.Password != "" {
		// 使用密码认证
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(proxyConfig.Password)}
	} else {
		return nil, fmt.Errorf("SSH代理需要提供密码或密钥")
	}

	// 连接到SSH服务器
	address := net.JoinHostPort(proxyConfig.Host, proxyConfig.Port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("连接SSH服务器失败: %w", err)
	}

	return &SSHProxy{
		client: client,
		config: &proxyConfig,
	}, nil
}

// Dial 通过SSH隧道建立到目标地址的连接
func (s *SSHProxy) Dial(network, address string) (net.Conn, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SSH客户端未初始化")
	}

	// 通过SSH隧道连接到目标地址
	conn, err := s.client.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("通过SSH隧道连接失败: %w", err)
	}

	return conn, nil
}

// Close 关闭SSH连接
func (s *SSHProxy) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// buildSSHProxyConfig 从ProxyConfig构建SSH代理配置JSON
func buildSSHProxyConfig(proxyConfig *database.ProxyConfig) (string, error) {
	sshConfig := SSHProxyConfig{
		Host:     proxyConfig.Host,
		Port:     proxyConfig.Port,
		User:     proxyConfig.User,
		Password: proxyConfig.Password,
		KeyFile:  proxyConfig.KeyFile,
	}

	// 如果提供了Config字段，尝试解析并合并
	if proxyConfig.Config != "" {
		var extraConfig map[string]interface{}
		if err := json.Unmarshal([]byte(proxyConfig.Config), &extraConfig); err == nil {
			// 合并额外配置
			if keyData, ok := extraConfig["key_data"].(string); ok {
				sshConfig.KeyData = keyData
			}
		}
	}

	configJSON, err := json.Marshal(sshConfig)
	if err != nil {
		return "", fmt.Errorf("序列化SSH配置失败: %w", err)
	}

	return string(configJSON), nil
}

