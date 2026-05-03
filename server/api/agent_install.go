package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"serversmith/db"

	"golang.org/x/crypto/ssh"
)

const (
	keyDir      = "/var/lib/serversmith/keys"
	privKeyPath = "/var/lib/serversmith/keys/serversmith_ed25519"
	pubKeyPath  = "/var/lib/serversmith/keys/serversmith_ed25519.pub"
)

// HandleSSHKey GET /api/ssh-key — 返回公钥内容（用于展示给用户复制）
func HandleSSHKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	pubKey, err := ensureKeyPair()
	if err != nil {
		writeError(w, 500, "key generation failed: "+err.Error())
		return
	}
	writeOK(w, map[string]string{"public_key": strings.TrimSpace(pubKey)})
}

// HandleAgentInstall POST /api/servers/:id/install-agent — SSH 推送安装探针
func HandleAgentInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	// 解析 server id
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// path: api/servers/{id}/install-agent
	if len(parts) < 4 {
		writeError(w, 400, "invalid path")
		return
	}
	var serverID int64
	if _, err := fmt.Sscanf(parts[2], "%d", &serverID); err != nil {
		writeError(w, 400, "invalid server id")
		return
	}

	// 查服务器信息
	var host, sshUser string
	var sshPort int
	err := db.DB.QueryRow(
		"SELECT host, ssh_port, ssh_user FROM servers WHERE id = ?", serverID,
	).Scan(&host, &sshPort, &sshUser)
	if err != nil {
		writeError(w, 404, "server not found")
		return
	}
	if sshPort == 0 {
		sshPort = 22
	}
	if sshUser == "" {
		sshUser = "root"
	}

	// 确保密钥对存在
	if _, err := ensureKeyPair(); err != nil {
		writeError(w, 500, "key error: "+err.Error())
		return
	}

	// 读私钥
	privKeyBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		writeError(w, 500, "cannot read private key")
		return
	}
	signer, err := ssh.ParsePrivateKey(privKeyBytes)
	if err != nil {
		writeError(w, 500, "cannot parse private key")
		return
	}

	// 建立 SSH 连接
	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 内网管理场景可接受
		Timeout:         30 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", host, sshPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		writeError(w, 502, "SSH connection failed: "+err.Error())
		return
	}
	defer client.Close()

	// 获取管理端地址（用于探针上报）
	managerHost := os.Getenv("MANAGER_PUBLIC_HOST")
	if managerHost == "" {
		managerHost = "89.58.2.247" // fallback，生产环境通过环境变量覆盖
	}
	managerPort := os.Getenv("MANAGER_PORT")
	if managerPort == "" {
		managerPort = "18080"
	}

	// 检测目标架构
	arch, err := runSSHCommand(client, "uname -m")
	if err != nil {
		writeError(w, 502, "cannot detect arch: "+err.Error())
		return
	}
	arch = strings.TrimSpace(arch)
	var goarch string
	switch arch {
	case "x86_64":
		goarch = "amd64"
	case "aarch64":
		goarch = "arm64"
	default:
		writeError(w, 400, fmt.Sprintf("unsupported arch: %s (only amd64/arm64)", arch))
		return
	}

	// 下载并安装探针的命令序列
	agentURL := fmt.Sprintf("http://%s:%s/agent/serversmith-agent-linux-%s", managerHost, managerPort, goarch)
	installCmd := fmt.Sprintf(`
set -e
mkdir -p /usr/local/bin
curl -sSL --max-time 30 "%s" -o /usr/local/bin/serversmith-agent
chmod +x /usr/local/bin/serversmith-agent
cat > /etc/systemd/system/serversmith-agent.service << 'EOF'
[Unit]
Description=ServerSmith Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/serversmith-agent --server http://%s:%s --server-id %d
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable serversmith-agent
systemctl restart serversmith-agent
echo "AGENT_INSTALLED_OK"
`, agentURL, managerHost, managerPort, serverID)

	output, err := runSSHCommand(client, installCmd)
	if err != nil {
		writeError(w, 502, "install failed: "+err.Error()+"\noutput: "+output)
		return
	}

	if !strings.Contains(output, "AGENT_INSTALLED_OK") {
		writeError(w, 502, "install may have failed, output: "+output)
		return
	}

	log.Printf("[agent-install] server %d (%s) agent installed successfully", serverID, host)
	writeOK(w, map[string]interface{}{
		"message": "探针安装成功",
		"arch":    arch,
		"output":  strings.TrimSpace(output),
	})
}

// HandleAgentBinary GET /agent/serversmith-agent-linux-{arch} — 提供探针下载
func HandleAgentBinary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	// 解析 arch
	base := filepath.Base(r.URL.Path)
	// base = serversmith-agent-linux-amd64 or serversmith-agent-linux-arm64
	var arch string
	if strings.HasSuffix(base, "amd64") {
		arch = "amd64"
	} else if strings.HasSuffix(base, "arm64") {
		arch = "arm64"
	} else {
		writeError(w, 400, "unknown arch")
		return
	}

	binPath := fmt.Sprintf("/var/lib/serversmith/agent/serversmith-agent-linux-%s", arch)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		writeError(w, 404, fmt.Sprintf("agent binary for %s not found on server, please upload it first", arch))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=serversmith-agent-linux-%s", arch))
	http.ServeFile(w, r, binPath)
}

// runSSHCommand 在已建立的 SSH 连接上执行命令，返回合并输出
func runSSHCommand(client *ssh.Client, cmd string) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}

// ensureKeyPair 确保 Ed25519 密钥对存在，返回公钥字符串
func ensureKeyPair() (string, error) {
	// 如果已存在直接返回
	if pubBytes, err := os.ReadFile(pubKeyPath); err == nil {
		return string(pubBytes), nil
	}

	// 生成新密钥对
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return "", err
	}

	// 用 ssh-keygen 生成（Go 标准库生成 ed25519 并序列化为 OpenSSH 格式）
	pubKey, privKey, err := generateED25519KeyPair()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(privKeyPath, privKey, 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(pubKeyPath, pubKey, 0644); err != nil {
		return "", err
	}

	log.Printf("[ssh-key] Generated new Ed25519 key pair at %s", keyDir)
	return string(pubKey), nil
}

// generateED25519KeyPair 生成 Ed25519 密钥对，返回 (公钥, 私钥, error)
func generateED25519KeyPair() ([]byte, []byte, error) {
	pubKey, privKey, err := generateKeyPairBytes()
	if err != nil {
		return nil, nil, err
	}
	return pubKey, privKey, nil
}

// HandleGetPublicKeyForServer GET /api/servers/:id/pubkey-install-cmd
// 返回给 Boss 复制的一条命令，粘贴到被管理服务器上即可写入公钥
func HandleGetPublicKeyForServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	pubKey, err := ensureKeyPair()
	if err != nil {
		writeError(w, 500, "key error: "+err.Error())
		return
	}
	pubKey = strings.TrimSpace(pubKey)
	installCmd := fmt.Sprintf(
		`mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo "公钥写入成功"`,
		pubKey,
	)
	writeOK(w, map[string]string{
		"public_key":  pubKey,
		"install_cmd": installCmd,
	})
}

// writeInstallJSON 辅助函数
func writeInstallJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
