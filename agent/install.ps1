# ServerSmith Agent 一键安装脚本 (Windows PowerShell)
# 用法: irm http://your-server:18080/agent/install.ps1 | iex
# 或指定参数: $env:MANAGER_URL="http://..."; $env:SERVER_ID="1"; irm .../install.ps1 | iex

param(
  [string]$ManagerUrl = $env:MANAGER_URL,
  [string]$ServerId   = $env:SERVER_ID
)

if (-not $ManagerUrl -or -not $ServerId) {
  Write-Host "用法: `$env:MANAGER_URL='http://your-server:18080'; `$env:SERVER_ID='1'; irm $ManagerUrl/agent/install.ps1 | iex"
  exit 1
}

$BinaryName = "serversmith-agent-windows-amd64.exe"
$InstallPath = "C:\serversmith\serversmith-agent.exe"
$ServiceName = "ServerSmithAgent"
$DownloadUrl = "$ManagerUrl/agent/$BinaryName"

Write-Host "[INFO] 从 $DownloadUrl 下载探针..."
New-Item -ItemType Directory -Force -Path "C:\serversmith" | Out-Null
Invoke-WebRequest -Uri $DownloadUrl -OutFile $InstallPath
Write-Host "[OK] 探针已下载到 $InstallPath"

# 注册 Windows 服务
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
  Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
  sc.exe delete $ServiceName | Out-Null
  Start-Sleep -Seconds 2
}

New-Service -Name $ServiceName `
  -BinaryPathName "`"$InstallPath`" --server $ManagerUrl --server-id $ServerId" `
  -DisplayName "ServerSmith Agent" `
  -StartupType Automatic `
  -Description "ServerSmith 服务器监控探针"

Start-Service -Name $ServiceName
Write-Host "[OK] Windows 服务已启动"
Write-Host ""
Write-Host "✅ ServerSmith 探针安装完成！约 30 秒后面板将显示此服务器在线。"
