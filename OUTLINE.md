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

## Instruções para IAs Futuras

- **Preservação de Idiomas Go**: Utilize sempre Go 1.26+, preferindo `any`, `slices.Contains`, `errors.Is` e outras modernizações.
- **Consistência Visual**: Mantenha a paleta VGA clássica definida em `internal/tui/theme.go`, preservando os perfis `default` (Borland blue) e `blue` (NC-style).
- **Consistência de Componentes**: Reutilize `dialogoOK` e `turboButton` para novos fluxos, evitando duplicação de desenho de diálogo/botão.
- **Menus e Atalhos**: Use sempre o sistema de `SetInputCapture` para garantir que atalhos de teclado (Alt+Hotkey, F10) funcionem em toda a aplicação. No dropdown, as hotkeys devem funcionar sem Alt.
