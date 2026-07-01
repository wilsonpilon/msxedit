# Projeto MSXEdit - Outline do Desenvolvimento

Este documento registra a filosofia, os prompts e o modo de pensar aplicados pela IA durante a criação do MSXEdit. Ele serve como referência para manter a consistência em futuras iterações ou outros projetos similares.

## Visão Geral do Projeto

O objetivo é criar um editor de texto estilo TUI (Text User Interface) focado em retrocomputação (MSX), rodando em terminais modernos (Go). A estética é baseada no Norton Editor e ferramentas da Borland (Turbo Vision).

## Modo de Pensar e Decisões de Arquitetura

1. **Linguagem de Implementação**: Go foi escolhido pela sua capacidade de gerar binários estáticos únicos, facilidade de manipulação de bytes (essencial para tokens BASIC) e excelentes bibliotecas de terminal (Tview/Tcell).
2. **Estrutura de Pacotes**:
   - `internal/cli`: Responsável por interpretar o mundo exterior (flags).
   - `internal/config`: Gerencia a persistência de preferências do usuário.
   - `internal/tui`: Separa a lógica de interface da lógica de negócio.
   - `internal/basic`: Focado na complexidade técnica de tokens e formatos binários do MSX.
3. **Prioridade de Configuração**: CLI > Config Local > Config Global > Padrões. Isso garante flexibilidade máxima para o usuário.
4. **Interface**: Uso de layouts Flex para garantir que o menu superior e a barra de status permaneçam fixos enquanto o editor ocupa o espaço restante.

## Prompts e Evolução

- **Início**: Estudo de viabilidade de syntax highlight em terminais modernos para linguagens retro.
- **Passo 2**: Definição do formato binário do MSX-BASIC (Header 0xFF, pointers, tokens).
- **Passo 3**: Implementação do sistema de CLI e Configuração.
- **Passo 4**: Criação do layout visual básico com Tview.
- **Passo 5**: Documentação e automação de build.
- **Passo 6**: Versionamento 4.0.3, build ID hexadecimal e menus interativos (File/Help).
- **Passo 7**: Redesign visual dos menus para o estilo Turbo Vision (cinza, branco, preto e vermelho).
- **Passo 8**: Ajuste fino das cores e bordas dos menus (fundo cinza, letras brancas, bordas brancas).
- **Passo 9**: Implementação do Desktop quadriculado, sombras nas janelas e estilo Turbo Pascal 7.0 exato.
- **Passo 10**: Redesign total para estilo monocromático (Preto/Cinza/Branco), removendo cores e fundos destacados para um visual limpo.
- **Passo 11**: Simplificação dos menus e implementação do sistema de teclas de atalho (Alt+Letra e Hotkeys em dropdowns) no estilo Turbo Vision.
- **Passo 12**: Introdução da janela de edição customizada no startup, com moldura dupla, título centralizado, número da janela e barras de rolagem horizontal/vertical renderizadas manualmente.
- **Passo 13**: Padronização de temas para paleta VGA explícita, com `default` = Borland blue (MS-DOS/Turbo) e `blue` = NC-style com menu/status ciano.
- **Passo 14**: Implementação do `Dialogo OK` reutilizável, com helper de centralização e suporte a botão customizável (label/hotkey/callback).
- **Passo 15**: Criação do componente `turboButton` e padronização visual dos botões com sombra estilo Turbo Vision.
- **Passo 16**: Adição de modos configuráveis de sombra (`shadowModeTurboClassic` e `shadowModeFlat`) para alternância rápida por tema/diálogo.
- **Passo 17**: Consolidação da janela de `Help` customizada, com moldura dupla, número de janela, barras de rolagem e comportamento independente do editor.
- **Passo 18**: Suporte a carregamento dinâmico de `HELP.md`, com parser de headings/links markdown e fallback interno quando o arquivo não estiver disponível.
- **Passo 19**: Navegação avançada no `Help` com breadcrumb, `Alt+F1`, fallback `Alt+Q`, subpainel especial para `Editor Commands` e seleção por teclado.
- **Passo 20**: Expansão do suporte a mouse para barra de menus, botões `[■]`, links do `Help` e trilhas de rolagem.
- **Passo 21**: Revisão completa da documentação para distinguir recursos já implementados dos itens ainda em evolução e consolidação do release `4.0.7`.
- **Passo 22**: Estudo completo do formato binário MSX-BASIC (projeto `basic-dignified`) e criação de `TOKEN.md` como referência definitiva de tokenização/detokenização.
- **Passo 23**: Expansão do menu `Options` com 12 itens reais (Compiler/Interpreter, Memory Sizes, Linker, Debugger, Directories, Tools, Environment►, Open, Save, Save as).
- **Passo 24**: Implementação do diálogo `Compiler/Interpreter Options` com 9 radio buttons em dois grupos, 3 checkboxes em bolinha, área de texto editável para defines condicionais e 3 botões Turbo Vision (OK/Cancel/Help).
- **Passo 25**: Syntax highlighting MSX-BASIC completo no editor — tokenizador `tokenizeBasicLine`, tabela de 100+ keywords, 11 categorias de token, suporte a todas as bases numéricas e zonas literais (REM/DATA/string/apostrophe).
- **Passo 26**: Windowing completo da janela de edição — arrastar pela barra de título, redimensionar pelo canto `◢`, maximizar/restaurar com `[▲]`/`[▼]`, scrollbars clicáveis, posição flutuante independente do layout flex, captura de mouse durante drag/resize.
- **Passo 27**: Criação do `msxread`, visualizador companheiro (segundo executável em `cmd/msxread`). Implementação do detokenizador MSX-BASIC funcional em `internal/basic` (`Detokenize`/`DetokenizeToText`, tabelas completas), pacote `internal/reader` (viewer de tela cheia: topo cinza, corpo cyan, status `Command►`) e CLI com `cobra`. Suporte a `.txt`, `.bas` tokenizado e `.md`.
- **Passo 28**: Expansão de recursos do `msxread`: syntax highlighting BASIC (keywords amarelo, strings branco, comentários cinza); ciclo de 16 cores VGA para texto/fundo (F5/F6/F7/F8); wrap de linha por fronteira de palavra (W), com `visualRow{line, col, len}` para corte limpo sem resíduos na linha anterior; hi-bit configurável (padrão ativo); busca interativa (F) com highlight em tempo real, próxima ocorrência (N), alternância case-sensitive (C); impressão (P); persistência de configurações em `msxread.json` ao lado do executável (S); overlay F1 com todas as teclas.
- **Passo 29**: Unificação do `build.ps1` versão 2.0. Flags `-Editor` e `-Reader` para compilar individualmente; sem flags compila os dois. Flags `-Run` (abre msxedit) e `-View` (abre msxread) pós-compilação. Build ID gerado uma única vez e injetado em ambos os binários via `-ldflags "-X main.BuildID=<hex>"`, garantindo que `msxedit --version` e `msxread --version` mostrem o mesmo ID de build. Versão unificada do sistema elevada para `4.1.5`.
- **Passo 30**: Aprimoramento da tela de ajuda do MSX-Read (`internal/reader/viewer.go`). Cabeçalho "Welcome to MSX-Read v:(versão)" centralizado na largura do painel (50 colunas), duas linhas em branco, "MSX-Read Help Screen" e "Copyright (c) 1972,2026 Cybernostra, Inc.". O campo `version` foi adicionado à struct `Viewer` e propagado de `cmd/msxread/main.go → reader.Options → NewViewer`.
- **Passo 31**: Implementação do diálogo **Open File** estilo Turbo Pascal no MSXEdit (`internal/tui/open_file_dialog.go`). Componente `openFileDialog` com desenho manual via `tview.Box`: campo `&Name` (fundo `vgaBlue`, texto `vgaLightCyan`), seta `↓` verde para histórico de máscaras, lista `&Files` bicolunar sem moldura (fundo `vgaCyan`), barra de rolagem horizontal azul (`◄▒■►`), área de status sem moldura (largura interna total) com caminho+máscara e detalhes do item selecionado, e quatro botões Turbo Vision de largura igual (`Open`, `Replace`, `Cancel`, `Help`). Ativado por `F3` globalmente via `SetInputCapture` e por `File → Open… F3` no menu. Função auxiliar `showOpenFileDialog(app, mask, onOpen, onReplace)`.
- **Passo 32**: Refinamento do layout do diálogo Open File. Linha em branco inserida entre o bloco Name/input e o bloco Files/lista; linha em branco entre o bloco Files/scrollbar e o bloco Status. Botões `Open` e `Replace` desceram 1 linha (alinhados ao campo de entrada e ao label `&Files`). `Cancel` e `Help` reposicionados acima da área de status. Correção do label `Cancel` de 8 para 7 caracteres plain — todos os quatro botões ficaram com largura idêntica de 11 colunas. Altura do diálogo ajustada para 19 linhas. Versão elevada para `4.1.7`.
- **Passo 33**: Virada de arquitetura de janela única para múltiplas janelas de edição. `a.Editors []*editorWindow` substitui o `editor` fixo de `a.Pages`; `createEditorWindow`/`closeEditorWindow`/`cascadePosition` (em `tui.go`) fabricam, posicionam em cascata e removem janelas, e `File → New` finalmente ganha ação real. Fechar a última janela deixou de encerrar o programa — `focusActiveEditor()` devolve o foco ao desktop quando não há nenhuma aberta. O clipboard deixou de ser por-janela (`w.blkClip`) e passou a ser compartilhado (`app.Clipboard`), com uma janela dedicada `Show clipboard` (`isClipboard bool`, borda amarela) sincronizada ao vivo por `syncClipboardWindow`.
- **Passo 34**: Implementação dos menus `Edit` e `Search` como fluxos reais, encerrando o estado de scaffolding que datava do Passo 11. `Edit`: `Undo`/`Redo` (delegando ao undo nativo do `tview.TextArea`), `Cut`/`Copy`/`Paste`/`Clear` (agora sobre o clipboard compartilhado do Passo 33). `Search`: diálogos `findDialog` (`find_dialog.go`), `replaceDialog` (`replace_dialog.go`) e `gotoLineDialog` (`goto_line_dialog.go`), todos construídos sobre o componente compartilhado `historyField` (`text_field.go`, campo com histórico por seta `↓`) e os helpers de desenho `drawGroupBox`/`drawCheckbox`/`drawRadio` (`dialog_widgets.go`). A lógica de busca em si (`findMatches`/`findNext`, suporte a case-sensitive, whole-words e regex) vive em `search.go`; `gotoTextLine`/`gotoBasicLine` (`goto_line.go`) resolvem números de linha de texto ou do MSX-Basic. O `Replace` já captura todas as opções e histórico, mas a substituição efetiva do texto ficou para uma iteração futura — documentado como limitação conhecida em vez de reivindicado como concluído.
- **Passo 35**: Editor ganhou o conjunto de atalhos estilo WordStar/Turbo que faltava para reivindicar paridade com o Norton Editor/Turbo Vision original: movimentação de cursor via `Ctrl+S/D/A/F/E/X/W/Z/R/C` (`selection.go`, `translateWordStarKey` traduz para as teclas de seta/página equivalentes e reaproveita toda a lógica de navegação), seleção de texto por `Shift`+navegação ou arraste do mouse (mesmo destaque `blk*` já usado pelos comandos de bloco), modo Insert/Overwrite, inserir/apagar linha, apagar até o fim da linha, apagar palavra à direita (`insert_delete.go`), place markers `Ctrl+K`/`Ctrl+Q 0-9`, restore line `Ctrl+Q L`, tab mode `Ctrl+O T`, auto indent `Ctrl+O I` e o prefixo `Ctrl+P` para códigos de controle literais (`misc_commands.go`). Os tópicos de Help "Cursor-movement commands" e "Insert & Delete commands" e "Miscelaneous commands" deixaram de ser placeholders — `help_window.go` generalizou `drawBlockCommandsPage` para `drawCommandsHeaderPage`, dirigido por um mapa `helpHeaderButtonTitles`, para desenhar o título em botão 3D em qualquer um desses tópicos, não só em "Block commands". De brinde, corrigido o erro de digitação histórico "Reserved Worlds" → "Reserved Words". Versão elevada para `4.1.9`.

## Instruções para IAs Futuras

- **Preservação de Idiomas Go**: Utilize sempre Go 1.26+, preferindo `any`, `slices.Contains`, `errors.Is` e outras modernizações.
- **Consistência Visual**: Mantenha a paleta VGA clássica definida em `internal/tui/theme.go`, preservando os perfis `default` (Borland blue) e `blue` (NC-style).
- **Consistência de Componentes**: Reutilize `dialogoOK` e `turboButton` para novos fluxos, evitando duplicação de desenho de diálogo/botão. Para diálogos com campos de texto/histórico, checkboxes ou radio buttons, reutilize `historyField` (`text_field.go`) e `drawGroupBox`/`drawCheckbox`/`drawRadio` (`dialog_widgets.go`) em vez de redesenhar do zero.
- **Menus e Atalhos**: Use sempre o sistema de `SetInputCapture` para garantir que atalhos de teclado (Alt+Hotkey, F10) funcionem em toda a aplicação. No dropdown, as hotkeys devem funcionar sem Alt.
- **Documentação Confiável**: Sempre documente separadamente o que já está funcional e o que ainda é scaffolding, placeholder ou roadmap, evitando promessas além do estado real do código.
