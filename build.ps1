param (
    [switch]$Windows,
    [switch]$Linux,
    [switch]$Release,
    [switch]$Dev,
    [switch]$Run,
    [Parameter(ValueFromRemainingArguments=$true)]
    $ExtraArgs
)

# Configurações padrão
$TargetOS = "windows"
if ($Linux) { $TargetOS = "linux" }

$Mode = "release"
if ($Dev) { $Mode = "debug" }

$OutputFile = "msxedit.exe"
if ($TargetOS -eq "linux") { $OutputFile = "msxedit" }

# Extrair versão do main.go
$Version = (Get-Content ./cmd/msxedit/main.go | Select-String 'const Version = "(.*)"').Matches.Groups[1].Value
if (-not $Version) { $Version = "Desconhecida" }

# Gerar Build ID em Hex (Unix UTC)
$BuildID = [Convert]::ToString([int64]([DateTimeOffset]::UtcNow.ToUnixTimeSeconds()), 16).ToUpper()

Write-Host "--- MSXEdit Build System ---" -ForegroundColor Cyan
Write-Host "Versão: $Version ($BuildID)" -ForegroundColor Green
Write-Host "Plataforma Alvo: $TargetOS"
Write-Host "Modo: $Mode"

# Atualizar dependências
Write-Host "Atualizando bibliotecas..." -ForegroundColor Gray
go mod tidy

# Definir flags de compilação
$BuildFlags = @()
if ($Mode -eq "release") {
    # Remover símbolos de debug e tabelas de símbolos para um binário menor
    $BuildFlags += "-ldflags"
    $BuildFlags += "-s -w"
}

# Configurar variáveis de ambiente para cross-compilation
$env:GOOS = $TargetOS
$env:GOARCH = "amd64"

# Compilar
Write-Host "Compilando..." -ForegroundColor Yellow
go build -o $OutputFile $BuildFlags ./cmd/msxedit/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build concluído com sucesso: $OutputFile" -ForegroundColor Green
    
    if ($Run) {
        if ($TargetOS -eq "windows") {
            Write-Host "Executando..." -ForegroundColor Cyan
            & ".\$OutputFile" $ExtraArgs
        } else {
            Write-Host "Aviso: Não é possível executar binários Linux nativamente no Windows sem WSL." -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "Erro durante a compilação." -ForegroundColor Red
    exit $LASTEXITCODE
}
