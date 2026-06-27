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

Ao iniciar, o programa cria automaticamente a primeira janela de edição (janela 1), mesmo sem arquivo informado.

### Teclas de Atalho

- **F1**: Ajuda
- **F2**: Salvar arquivo
- **F3**: Abrir arquivo
- **F10 / ESC**: Sair do programa
- **ALT + Primeira Letra do Menu**: Acessar menus superiores (ex: ALT+F para File)

### Interface

1. **Barra de Menu (Topo)**: Navegação estilo Top-Down para acesso a funções de arquivo, edição e opções.
2. **Área de Edição (Centro)**: Onde o código é escrito. A janela usa moldura branca de linhas duplas e fundo azul VGA clássico (Borland), com título centralizado, identificador de janela e indicadores visuais no topo.
3. **Barra de Status (Base)**: Exibe informações sobre o cursor, o arquivo aberto e atalhos rápidos.
4. **Barras de Rolagem (Moldura da Janela)**: Horizontal na base da janela e vertical na lateral direita, com setas, trilho quadriculado e cursor em bloco.

### Diálogos

- O projeto possui o componente reutilizável **Dialogo OK** (`dialogoOK`) para janelas de confirmação/aviso.
- O helper `showDialogoOKCentered(...)` centraliza o diálogo na tela inteira com foco automático.
- O botão principal do diálogo pode ser configurado via `SetButton(label, hotkey, callback)`.

### Botões (estilo Turbo)

- O componente `turboButton` é reutilizável e usado pelo `Dialogo OK`.
- Modos de sombra suportados:
  - `shadowModeTurboClassic`: sombra clássica estilo Turbo Vision (`[BOTAO]▄` + linha inferior deslocada).
  - `shadowModeFlat`: sombra inferior reta para estilo mais discreto.
- A troca rápida por diálogo pode ser feita com `SetButtonShadowMode(...)`.

### Temas de Cores

- **`default`**: VGA Borland blue (MS-DOS/Turbo style).
- **`blue`**: VGA NC-style (Norton Commander), com barra de menu e status em ciano.

## Configuração

O programa busca configurações na seguinte ordem:
1. Argumentos de linha de comando.
2. Arquivo `msxedit.json` no diretório local (se o parâmetro `--local` for usado).
3. Arquivo `msxedit.json` no diretório de configuração do usuário (`%APPDATA%/msxedit/` no Windows ou `~/.config/msxedit/` no Linux).

Para mais detalhes sobre as opções, consulte o arquivo [REFERENCE.md](REFERENCE.md).
