param (
    [switch]$Windows,   # força alvo Windows (comportamento padrão)
    [switch]$Linux,     # gera binário Linux (cross-compile)
    [switch]$Release,   # flags de release — comportamento padrão
    [switch]$Dev,       # compila sem stripping de símbolos
    [switch]$Editor,    # compila apenas msxedit
    [switch]$Reader,    # compila apenas msxread
    [switch]$Run,       # abre msxedit após compilar
    [switch]$View,      # abre msxread após compilar
    [Parameter(ValueFromRemainingArguments=$true)]
    $ExtraArgs
)

# ── Plataforma e modo ────────────────────────────────────────────────────────
$TargetOS = if ($Linux) { "linux" } else { "windows" }
$Mode     = if ($Dev)   { "debug" } else { "release" }
$Ext      = if ($TargetOS -eq "windows") { ".exe" } else { "" }

$EditorOutput = "msxedit$Ext"
$ReaderOutput = "msxread$Ext"

# ── O que compilar ───────────────────────────────────────────────────────────
# Sem -Editor nem -Reader → compila os dois.
# -Run implica que msxedit esteja disponível; -View implica msxread.
if (-not $Editor -and -not $Reader) {
    $BuildEditor = $true
    $BuildReader  = $true
} else {
    $BuildEditor = $Editor.IsPresent
    $BuildReader  = $Reader.IsPresent
}
# Avisar se o usuário pedir -Run/-View mas não tiver compilado o alvo
if ($Run  -and -not $BuildEditor) {
    Write-Host "Aviso: -Run especificado mas msxedit não será compilado (-Editor não está ativo)." -ForegroundColor Yellow
}
if ($View -and -not $BuildReader) {
    Write-Host "Aviso: -View especificado mas msxread não será compilado (-Reader não está ativo)." -ForegroundColor Yellow
}

# ── Versão e Build ID compartilhados ────────────────────────────────────────
# A versão é lida do msxedit; ambos os binários recebem o mesmo valor.
# O Build ID é gerado aqui (compile-time) e injetado via -ldflags para que
# msxedit --version e msxread --version mostrem exatamente o mesmo ID.
$Version = (Get-Content ./cmd/msxedit/main.go |
            Select-String 'const Version = "(.*)"').Matches.Groups[1].Value
if (-not $Version) { $Version = "Desconhecida" }

$BuildID = [Convert]::ToString(
    [int64]([DateTimeOffset]::UtcNow.ToUnixTimeSeconds()), 16
).ToUpper()

# ── Cabeçalho ────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "┌─ MSXEdit Build System ──────────────────────────────┐" -ForegroundColor Cyan
Write-Host "│  Versão : $Version                                      │" -ForegroundColor Cyan
Write-Host "│  Build  : $BuildID                                     │" -ForegroundColor Cyan
Write-Host "│  Alvo   : $TargetOS  |  Modo: $Mode                    │" -ForegroundColor Cyan
Write-Host "└────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

# ── Dependências ─────────────────────────────────────────────────────────────
Write-Host "Atualizando bibliotecas..." -ForegroundColor Gray
go mod tidy

# ── Flags de compilação ───────────────────────────────────────────────────────
# -X injeta BuildID nos dois binários em tempo de compilação.
$LdFlags = "-X main.BuildID=$BuildID"
if ($Mode -eq "release") { $LdFlags += " -s -w" }

$env:GOOS   = $TargetOS
$env:GOARCH = "amd64"

# ── Compilar msxedit ─────────────────────────────────────────────────────────
if ($BuildEditor) {
    Write-Host "Compilando msxedit..." -ForegroundColor Yellow
    go build -o $EditorOutput -ldflags $LdFlags ./cmd/msxedit/main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Erro ao compilar msxedit." -ForegroundColor Red
        exit $LASTEXITCODE
    }
    Write-Host "  → $EditorOutput  OK" -ForegroundColor Green
}

# ── Compilar msxread ─────────────────────────────────────────────────────────
if ($BuildReader) {
    Write-Host "Compilando msxread..." -ForegroundColor Yellow
    go build -o $ReaderOutput -ldflags $LdFlags ./cmd/msxread/main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Erro ao compilar msxread." -ForegroundColor Red
        exit $LASTEXITCODE
    }
    Write-Host "  → $ReaderOutput  OK" -ForegroundColor Green
}

Write-Host ""
Write-Host "Build concluído: versão $Version  build $BuildID" -ForegroundColor Green

# ── Executar após compilar ───────────────────────────────────────────────────
if ($Run) {
    if ($TargetOS -eq "windows") {
        Write-Host "Executando msxedit..." -ForegroundColor Cyan
        & ".\$EditorOutput" $ExtraArgs
    } else {
        Write-Host "Aviso: binário Linux não pode ser executado no Windows sem WSL." -ForegroundColor Yellow
    }
}

if ($View) {
    if ($TargetOS -eq "windows") {
        Write-Host "Executando msxread..." -ForegroundColor Cyan
        & ".\$ReaderOutput" $ExtraArgs
    } else {
        Write-Host "Aviso: binário Linux não pode ser executado no Windows sem WSL." -ForegroundColor Yellow
    }
}
