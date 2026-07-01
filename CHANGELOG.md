# Changelog

Todas as mudancas relevantes deste projeto serao documentadas neste arquivo.

Formato baseado em categorias (`Added`, `Changed`, `Fixed`, `Docs`).

## [4.1.9] - 2026-07-01

### Added
- **Múltiplas janelas de edição** no MSXEdit: `File → New` (`internal/tui/tui.go`) cria uma nova
  janela em branco em cascata a partir da janela ativa, com ID próprio, clipboard compartilhado
  entre todas as janelas e ativação automática ao ganhar foco.
  - `Alt+F3` fecha a janela de edição ativa; fechar a última janela aberta **não** encerra mais o
    programa — o desktop quadriculado fica visível e `File → New`/`Open…` continuam disponíveis.
  - As janelas de edição/ajuda/clipboard agora vivem como páginas independentes de `a.Pages`
    sobre o desktop, em vez de uma única página fixa.
- **Menu `Edit` funcional** (`internal/tui/menu.go`): `Undo` (`Alt+BkSp`), `Redo` (só pelo menu —
  `Ctrl+Y` já é usado por Delete line), `Cut` (`Shift+Del`), `Copy` (`Ctrl+Ins`), `Paste`
  (`Shift+Ins`), `Clear` (`Ctrl+Del`) e `Show clipboard`.
  - **Clipboard compartilhado** (`app.Clipboard`) entre todas as janelas de edição.
  - **Janela "Show clipboard"**: exibe/edita o clipboard compartilhado ao vivo, com borda amarela
    distintiva; qualquer Copy/Cut/Paste em qualquer janela reflete nela imediatamente.
- **Menu `Search` funcional** (`internal/tui/menu.go`, `search.go`, `find_dialog.go`,
  `replace_dialog.go`, `goto_line.go`, `goto_line_dialog.go`):
  - **`Find...`**: diálogo com campo de texto com histórico (seta `↓`), opções `Case sensitive`,
    `Whole words only`, `Regular expression`, grupos `Direction` (Forward/Backward), `Scope`
    (Global/Selected text) e `Origin` (From cursor/Entire scope). Busca real no editor ativo,
    seleciona a ocorrência encontrada e avisa na barra de status quando o texto não é encontrado
    ou quando a busca reinicia do começo (wrap).
  - **`Search again`** (também `Ctrl+L`): repete a última busca a partir da posição atual do
    cursor; sem busca anterior, abre o diálogo `Find`.
  - **`Go to line number...`**: diálogo com campo numérico e opção "linha do MSX-Basic" (procura
    pelo número de linha do BASIC no início da linha, ex.: `10 PRINT...`) versus linha de texto
    (1-based). Move o cursor e avisa na barra de status se a linha não existir.
  - **`Replace...`**: diálogo completo (texto a buscar, novo texto, `Prompt on replace`, mesmas
    opções/grupos do `Find`, botões `Ok`/`Change all`/`Cancel`/`Help`) que já captura todas as
    opções e o histórico de ambos os campos — a substituição efetiva do texto ainda não está
    implementada (ver `Docs`/limitações).
  - `Show last compiler error`, `Find error...` e `Find procedure...` permanecem placeholders.
- **Seleção de texto com Shift e mouse** (`internal/tui/selection.go`): `Shift`+setas/`Home`/`End`/
  `PgUp`/`PgDn` estende uma seleção a partir de um ponto-âncora; arrastar o mouse com o botão
  esquerdo pressionado faz o mesmo. Ambas reaproveitam o destaque e os comandos de bloco/clipboard
  já existentes (`Ctrl+K C/V/Y`, `Ctrl+Ins`, `Shift+Del`, `Shift+Ins`, `Ctrl+Del`).
- **Atalhos de movimentação de cursor estilo WordStar/Turbo** (`internal/tui/selection.go`):
  `Ctrl+S/D` (caractere esq./dir.), `Ctrl+A/F` (palavra esq./dir.), `Ctrl+E/X` (linha cima/baixo),
  `Ctrl+W/Z` (rolar tela cima/baixo sem mover o cursor), `Ctrl+R/C` (página cima/baixo) e
  `Ctrl+Up`/`Ctrl+Down` (meia tela cima/baixo).
- **Comandos de inserção/remoção** (`internal/tui/insert_delete.go`): modo `Insert`/`Overwrite`
  (`Ctrl+V` ou `Ins` sem modificador alterna), `Ctrl+N` (insere linha em branco), `Ctrl+Y` (remove
  a linha inteira), `Ctrl+Q Y` (remove até o fim da linha), `Ctrl+G` (apaga caractere à direita,
  equivalente a `Del`), `Ctrl+T` (apaga palavra à direita).
- **Comandos diversos** (`internal/tui/misc_commands.go`):
  - **Place markers** `Ctrl+K 0-9` (grava posição) / `Ctrl+Q 0-9` (vai até a posição gravada).
  - **Restore line** (`Ctrl+Q L`): desfaz todas as edições feitas na linha atual desde que o
    cursor chegou nela, independentemente do histórico normal de Undo.
  - **Tab mode** (`Ctrl+O T`): alterna entre inserir caractere de tabulação (padrão) ou espaços
    até a próxima parada de 8 colunas.
  - **Auto indent** (`Ctrl+O I`): ao pressionar Enter, repete a indentação (espaços/tabs iniciais)
    da linha anterior.
  - **Ctrl+character prefix** (`Ctrl+P`): insere o byte de controle (1–26) correspondente à
    próxima tecla `Ctrl+letra` pressionada, no estilo WordStar/Turbo.
  - **`Ctrl+K S`**: salva o arquivo sem sair do editor (mesmo fluxo de `F2`).
  - **`Ctrl+F1`** ("Language help"): abre o `Help` já navegado até o tópico `Reserved Words`.
- **Undo/Redo** (`internal/tui/block_commands.go`): `Edit → Undo`/`Alt+BkSp` e `Edit → Redo`
  disparam o undo/redo nativo do `tview.TextArea`.
- Novos tópicos de `Help` com conteúdo real, deixando de ser placeholders: **Cursor-movement
  commands**, **Insert & Delete commands** e **Miscelaneous commands** — todos com título em
  botão 3D, no mesmo estilo já usado por **Block commands** (`helpHeaderButtonTitles` em
  `internal/tui/help_window.go` generaliza `drawBlockCommandsPage` para `drawCommandsHeaderPage`).

### Changed
- **`Ctrl+K`** ganhou `S` (Save) e `0-9` (set marker); barra de status do prefixo atualizada.
- **`Ctrl+Q`** ganhou `Y` (delete to EOL), `L` (restore line), `F` (Find), `A` (Replace) e `0-9`
  (goto marker); barra de status do prefixo atualizada.
- **Novos prefixos `Ctrl+O`** (Tab mode / Auto indent) e **`Ctrl+P`** (Ctrl+character literal).
- **`File → New`** deixou de ser um item inerte e agora cria uma janela de edição real.
- Fechar o menu (`Esc` ou clique fora) e fechar o `Help` agora devolvem o foco para a **janela de
  edição ativa** (`a.focusActiveEditor()`), em vez de sempre para `a.Editor` (que passa a
  referenciar apenas a janela ativa mais recente, não uma janela fixa única).
- Corrigido o texto do tópico de Help **"Reserved Worlds" → "Reserved Words"** (erro de digitação)
  em `internal/tui/help_content.go` e `HELP.md`.
- Versão geral elevada para `4.1.9` em ambos os binários (`msxedit` e `msxread`).

### Docs
- `README.md`, `MANUAL.md`, `REFERENCE.md`, `OUTLINE.md`, `HELP.md` atualizados com múltiplas
  janelas de edição, menus `Edit`/`Search` funcionais, clipboard compartilhado, seleção de texto,
  novos atalhos de edição/navegação estilo WordStar e a nova versão.
- Registrado como limitação conhecida: o diálogo `Replace` já captura texto/opções/histórico, mas
  ainda não executa a substituição no texto do editor.

## [4.1.7] - 2026-06-29

### Added
- **Diálogo Open File (F3) no MSXEdit** (`internal/tui/open_file_dialog.go`): janela modal
  estilo Turbo Pascal com layout em blocos bem definidos:
  - Campo `&Name` com fundo azul escuro (`vgaBlue`) e texto em ciano claro, seta `↓` verde
    para histórico de máscaras (padrão `*.BAS`).
  - Lista `&Files` bicolunar em ciano (`vgaCyan`), sem moldura, com divider `│` e coluna de
    diretórios; 8 linhas visíveis.
  - Barra de rolagem horizontal em azul com controles `◄▒■►` em ciano claro.
  - Área de status completa (sem moldura, largura interna total): linha 1 = caminho+máscara;
    linha 2 = nome do item, tipo (file/directory), data, hora e tamanho.
  - Quatro botões Turbo Vision de largura igual (11 colunas): `Open`, `Replace`, `Cancel`, `Help`.
  - Ativado por `F3` globalmente e por `File → Open… F3` no menu.
- **Dropdown de histórico de máscaras** (`drawHistoryDropdown`): lista com moldura que aparece
  abaixo do campo `&Name` ao pressionar a seta `↓`.
- **Cabeçalho da tela de ajuda do MSX-Read** (`internal/reader/viewer.go`):
  - Linha "Welcome to MSX-Read v:(versão)" centralizada na largura do painel.
  - Duas linhas em branco de separação.
  - "MSX-Read Help Screen" e "Copyright (c) 1972,2026 Cybernostra, Inc.".
  - Lista completa de atalhos a seguir.

### Changed
- **Layout do diálogo Open File** refinado:
  - Linha em branco entre o bloco Name/input e o bloco Files/lista.
  - Linha em branco entre o bloco Files/scrollbar e o bloco Status.
  - `Open` e `Replace` descem 1 linha para alinhar com o campo de entrada e o label `&Files`.
  - `Cancel` e `Help` reposicionados acima da área de status.
  - Altura do diálogo ajustada para 19 linhas (terminando logo após o bloco de status).
  - Todos os quatro botões com mesmo texto plain de 7 caracteres → largura 11 idêntica.
- Versão geral elevada para `4.1.7` em ambos os binários (`msxedit` e `msxread`).

### Docs
- `README.md`, `MANUAL.md`, `REFERENCE.md`, `OUTLINE.md` atualizados com o diálogo Open File,
  melhorias do MSX-Read e nova versão.
- Imagem `images/MSX-Read-01.png` incluída no README.

## [4.1.5] - 2026-06-29

### Added
- **`msxread`** — novo utilitário visualizador de textos, par do `msxedit`, no estilo do
  leitor de `README` do Turbo Pascal. Executável independente (`cmd/msxread`):
  - Exibe **3 tipos de arquivo**: `.txt` (texto puro), `.bas` **tokenizado** (detokenizado
    para listagem BASIC legível) e `.md` (ajuda, render leve com títulos e links).
  - **Barra de topo** cinza com letras pretas: data, hora e nome do arquivo separados por
    losango `◆`.
  - **Corpo** com fundo cyan e letras pretas, rolável.
  - **Barra de status** cinza: `Command►` + indicador de posição (`*** Top of File ***`,
    `*** End of File ***` ou `Line N of M`) + mini-help `Keys: ↑ ↓ ← → PgUp PgDn  ESC=Exit  F1=Help`.
  - Navegação: setas, `PgUp`/`PgDn`, `Home`/`End`, roda do mouse; `F1` abre overlay de ajuda;
    `ESC` (ou `Q`) sai.
  - Relógio ao vivo (atualiza a hora no topo a cada segundo).
  - CLI com **cobra**: argumento posicional `<arquivo>`, flags `--type/-t (auto|txt|bas|md)`,
    `--tabsize`, `--version/-v`, `--help/-h`.
- **Detokenizador MSX-BASIC** funcional em `internal/basic`:
  - `internal/basic/detokenize.go`: `IsTokenized`, `Detokenize`, `DetokenizeToText`.
  - Tabelas completas de tokens simples (0x81–0xFC) e estendidos (prefixo 0xFF) em `tokens.go`.
  - Decodificação de números (dígito imediato, byte, word, ref. de linha, `&H`, `&O`, float
    single/double BCD) e zonas literais (`REM`, `DATA`, `CALL`, strings, `'`, `ELSE`).
  - Testes (`detokenize_test.go`) cobrindo os exemplos de `TOKEN.md` seção 9.

### Changed
- `build.ps1` agora compila **ambos** os binários (`msxedit` e `msxread`) e reporta as duas versões.
- `internal/basic/tokens.go`: stub `Processor`/`LoadTokenized` substituído pelas tabelas
  completas usadas pelo detokenizador.

### Docs
- `README.md`, `MANUAL.md`, `REFERENCE.md`, `OUTLINE.md` atualizados com o utilitário `msxread`.

## [4.1.0] - 2026-06-29

### Added
- **Syntax highlighting** para MSX-BASIC no editor:
  - Novo arquivo `msxbasic_highlight.go` com tokenizador completo (`tokenizeBasicLine`)
  - Categorias: número de linha, statement, modifier, função built-in, string, número, comentário, variável, operador, símbolo
  - Mais de 100 keywords mapeadas na tabela `msxKeywordMap`
  - Números em todas as bases: decimal, `&H` (hex), `&O` (octal), `&B` (binário), com sufixos de tipo
  - Zonas literais: `REM …`, `DATA …`, strings `"…"`, comentário `'…`
  - Campo `highlightEnabled` em `editorWindow` controla ativação por janela
- **Janela de edição flutuante** com windowing completo:
  - Arrastar pela barra de título (clique e movimento)
  - Redimensionar pelo canto inferior direito (`◢`)
  - Botão de zoom `[▲]`/`[▼]` para maximizar/restaurar
  - Barras de rolagem horizontal e vertical clicáveis
  - `MouseScrollUp`/`Down` rola o conteúdo
  - Captura de mouse durante drag/resize via `return (true, w)` no `MouseHandler`
  - Posicionamento e tamanho armazenados em `winX/winY/winW/winH` (independentes do layout flex)
- **Diálogo Compiler/Interpreter Options** (`compiler_options_dialog.go`):
  - 9 opções de linguagem em dois grupos: Basic Code (MSX-BASIC, Basic Dignified, MSX Bas2Rom, Turbo Basic, NBasic) e Others (MSXgl/SDCC, N80/LK80, ASCII-C, Turbo Pascal 3.3f)
  - 3 checkboxes: Extended syntax, Overflow checking, Strict vars
  - Área de texto "Conditional defines:" com cursor e edição completa (insert, delete, Enter para nova linha)
  - 3 botões estilo Turbo Vision: OK, Cancel, Help (com destaque de foco)
  - Navegação por Tab, Shift+Tab, setas, Space, hotkeys O/C/H
  - Atalho `Help` abre diálogo de ajuda secundário (`showCompilerOptionsHelp`)
- **Menu Options** expandido com 12 itens:
  - Compiler/Interpreter… (ativo — abre diálogo)
  - Memory Sizes…, Linker…, Debugger…, Directories, Tools (scaffold)
  - Environment ► (submenu futuro)
  - Open…, Save …\msxedit.cfg, Save as… (scaffold)
- **TOKEN.md**: referência completa do formato binário MSX-BASIC (tabela de tokens simples 0x81–0xFC, tokens estendidos 0xFF+byte, encoding de números, estrutura de linha, algoritmos de tokenização/detokenização com exemplos hexadecimais anotados)

### Changed
- `editorWindow` reescrita para windowing flutuante: ignora rect do layout flex, usa `winX/winY/winW/winH` próprios; inicializa posição na primeira chamada a `Draw()`
- `tui.go`: janela do editor adicionada diretamente ao `pages` (sem wrapper `editorHost` Flex), permitindo posicionamento livre na tela
- Canto inferior direito da janela de edição mudou de `╝` para `◢` (handle de resize)
- Botão de zoom mudou de `[↕]` para `[▲]`/`[▼]` (indica estado atual)

### Docs
- `TOKEN.md` criado como referência definitiva de tokenização MSX-BASIC
- `CHANGELOG.md`, `OUTLINE.md`, `REFERENCE.md` atualizados para refletir o estado da release 4.1.0

## [4.0.7] - 2026-06-27

### Added
- Janela de `Help` navegavel com conteudo carregado de `HELP.md`, fallback interno e topicos interligados por links.
- Navegacao por breadcrumb no `Help`, com retorno por `Alt+F1` e fallback `Alt+Q` para terminais que nao entregam a combinacao original.
- Suporte a mouse na interface principal para:
  - clique na barra de menus
  - clique no controle `[■]` para fechar janelas/dialogos
  - clique em links e barras de rolagem da janela de `Help`
- Subjanela centralizada para `Editor Commands` dentro do `Help`, com cabecalho estilo Turbo e lista selecionavel.

### Changed
- Documentacao principal sincronizada com o estado real do projeto:
  - `README.md`
  - `MANUAL.md`
  - `HELP.md`
  - `REFERENCE.md`
  - `OUTLINE.md`
- Menus documentados de forma mais precisa: `File` e `Help` possuem fluxos ativos; os demais itens permanecem como scaffolding visual para iteracoes futuras.
- A release-base do projeto passa a ser identificada como `4.0.7`.

### Fixed
- Referencias desatualizadas de versao (`4.0.3`) substituidas por `4.0.7` onde representavam a versao corrente.
- Documentacao corrigida para separar recursos implementados de recursos ainda planejados (como `Open/Save`, `Compile/Make`, `syntax highlight` completo e parser tokenizado de BASIC).

### Docs
- Ajuda embarcada reescrita para refletir o comportamento atual da UI, atalhos, temas, mouse, menus e limitacoes conhecidas.
- Manual e referencia atualizados com parametros reais de build/configuracao e observacoes sobre funcionalidades em andamento.

## [4.0.3] - 2026-06-27

### Added
- Janela de edicao inicial customizada no startup, com suporte a multiplas janelas na estrutura interna (`Editors` e `ActiveEditor`).
- Renderizacao de janela estilo Turbo Vision com moldura dupla, titulo centralizado, identificador de janela e controles no topo.
- Barras de rolagem customizadas na moldura (horizontal e vertical), com setas, trilho e cursor visual.
- Componente reutilizavel `dialogoOK` para dialogos de confirmacao/informacao.
- Helper de layout `showDialogoOKCentered(...)` para centralizar qualquer `dialogoOK`.
- Componente reutilizavel `turboButton` para botao estilo Turbo Vision.
- Configuracao de sombra do botao por modo:
  - `shadowModeTurboClassic`
  - `shadowModeFlat`

### Changed
- Tema `default` migrado para paleta VGA explicita com foco em Borland blue classico (`RGB 0,0,170`).
- Tema `blue` padronizado para usar apenas a mesma tabela VGA interna (sem mistura com `tcell.Color*`).
- Dialogo `About` refeito como dialogo custom:
  - fundo cinza
  - moldura dupla branca
  - controle `[■]` no topo esquerdo com quadrado verde
  - botao `O&K` verde com texto branco e hotkey em amarelo
- `About` agora usa o componente `dialogoOK` e abertura centralizada.
- Menus `Help` atualizados com novos subitens e separadores conforme especificacao.

### Fixed
- Ajustes progressivos de alinhamento e orientacao da sombra 3D do botao para o padrao visual esperado do estilo Turbo.
- Correcao de import nao utilizado apos refatoracoes do dialogo `About`.

### Docs
- Atualizacao de `README.md`, `MANUAL.md`, `REFERENCE.md` e `OUTLINE.md` para refletir:
  - temas VGA (`default`, `blue`)
  - novos componentes de UI reutilizaveis
  - evolucao da arquitetura e padroes visuais
- Versao de referencia atualizada para `4.0.3` na documentacao da epoca.

## [4.0.1] - 2026-06-xx

### Added
- Base de menus interativos (`File/Help`) e hotkeys principais.
- Estrutura inicial de build/versionamento documentada.

### Notes
- Esta secao representa a linha-base historica antes das mudancas detalhadas em `4.0.3`.

