# Manual de Operação e Compilação - MSXEdit

Este documento descreve como compilar, configurar e operar o MSXEdit.

## Compilação

O MSXEdit é escrito em Go e requer o Go 1.26 ou superior instalado em seu sistema.

### Usando o Script de Build (Recomendado)

No Windows, você pode usar o script `build.ps1` para facilitar o processo:

```powershell
# Compilar para Windows (padrão, Release)
.\build.ps1

# Compilar para Linux
.\build.ps1 -Linux

# Compilar em modo Dev e executar
.\build.ps1 -Dev -Run
```

### Compilação Manual

```bash
# Baixar dependências
go mod tidy

# Compilar
go build -o msxedit.exe ./cmd/msxedit/main.go
```

## Operação

O MSXEdit pode ser iniciado via linha de comando. Se um nome de arquivo for fornecido, ele será aberto para edição.

### Teclas de Atalho

- **F1**: Ajuda
- **F2**: Salvar arquivo
- **F3**: Abrir arquivo
- **F10 / ESC**: Sair do programa
- **ALT + Primeira Letra do Menu**: Acessar menus superiores (ex: ALT+F para File)

### Interface

1. **Barra de Menu (Topo)**: Navegação estilo Top-Down para acesso a funções de arquivo, edição e opções.
2. **Área de Edição (Centro)**: Onde o código é escrito. Possui suporte a syntax highlight automático baseado na extensão do arquivo.
3. **Barra de Status (Base)**: Exibe informações sobre o cursor, o arquivo aberto e atalhos rápidos.

## Configuração

O programa busca configurações na seguinte ordem:
1. Argumentos de linha de comando.
2. Arquivo `msxedit.json` no diretório local (se o parâmetro `--local` for usado).
3. Arquivo `msxedit.json` no diretório de configuração do usuário (`%APPDATA%/msxedit/` no Windows ou `~/.config/msxedit/` no Linux).

Para mais detalhes sobre as opções, consulte o arquivo [REFERENCE.md](REFERENCE.md).
