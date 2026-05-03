package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const agentDir = "/var/lib/serversmith/agent"

// HandleAgentInstallScript GET /agent/install.sh  GET /agent/install.ps1
func HandleAgentInstallScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	name := filepath.Base(r.URL.Path) // install.sh or install.ps1
	scriptPath := filepath.Join(agentDir, name)

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		writeError(w, 404, "install script not found")
		return
	}

	if strings.HasSuffix(name, ".ps1") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/x-sh; charset=utf-8")
	}
	http.ServeFile(w, r, scriptPath)
}

// HandleAgentBinary GET /agent/serversmith-agent-{platform}-{arch}[.exe]
func HandleAgentBinary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	name := filepath.Base(r.URL.Path)
	// 安全检查：只允许下载 serversmith-agent-* 文件
	if !strings.HasPrefix(name, "serversmith-agent-") {
		writeError(w, 400, "invalid binary name")
		return
	}

	binPath := filepath.Join(agentDir, name)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		writeError(w, 404, fmt.Sprintf("agent binary '%s' not found on server", name))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name))
	http.ServeFile(w, r, binPath)
}

// HandleSSHKey GET /api/ssh-key — 返回管理端公钥（供 Boss 查看）
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

// HandleGetPublicKeyForServer GET /api/servers/:id/pubkey-cmd
// 返回一条命令，粘贴到被管理服务器上写入公钥（用于 SSH 重启功能）
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

// ensureAgentDir 确保 agent 目录存在并同步脚本和二进制
func EnsureAgentAssets() {
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		log.Printf("[agent] Failed to create agent dir: %v", err)
		return
	}

	// 把编译好的二进制从 /usr/local/bin 同步过来（如果存在）
	// 生产部署时直接上传到 agentDir 即可
	log.Printf("[agent] Agent assets dir: %s", agentDir)
}
