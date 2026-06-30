package tui

import (
	"strings"
)

// LinkInfo armazena informações sobre um link no Help
type LinkInfo struct {
	Text    string // Texto do link
	TopicID string // ID do tópico para o qual o link aponta
	Row     int    // Linha em que o link aparece
	ColStart int   // Coluna inicial do link
	ColEnd  int    // Coluna final do link
}

// HelpTopic representa um tópico de ajuda
type HelpTopic struct {
	ID    string     // ID único do tópico
	Title string     // Título do tópico
	Lines []string   // Linhas do conteúdo
	Links []LinkInfo // Links neste tópico
}

// HelpContent gerencia todo o conteúdo do Help
type HelpContent struct {
	topics         map[string]*HelpTopic
	currentTopicID string
	history        []string // Histórico de navegação
}

// NewHelpContent cria um novo gerenciador de conteúdo de Help
func NewHelpContent() *HelpContent {
	hc := &HelpContent{
		topics:         make(map[string]*HelpTopic),
		currentTopicID: "contents",
		history:        []string{},
	}
	if err := hc.loadTopicsFromMarkdown(); err != nil {
		hc.initializeTopics()
	}
	return hc
}

// linkDef define um link a ser localizado nas linhas do tópico
type linkDef struct {
	Text    string
	TopicID string
}

// buildLinks encontra automaticamente as posições dos links nas linhas
func buildLinks(lines []string, defs []linkDef) []LinkInfo {
	links := []LinkInfo{}
	for _, def := range defs {
		for rowIdx, line := range lines {
			runes := []rune(line)
			textRunes := []rune(def.Text)
			// Procura a sequência de runes no rune slice da linha
			for col := 0; col <= len(runes)-len(textRunes); col++ {
				match := true
				for i, r := range textRunes {
					if runes[col+i] != r {
						match = false
						break
					}
				}
				if match {
					links = append(links, LinkInfo{
						Text:     def.Text,
						TopicID:  def.TopicID,
						Row:      rowIdx,
						ColStart: col,
						ColEnd:   col + len(textRunes),
					})
					break // Pega apenas a primeira ocorrência na linha
				}
			}
		}
	}
	return links
}

// initializeTopics inicializa todos os tópicos de Help
func (hc *HelpContent) initializeTopics() {
	// ── Tópico Contents (índice) ─────────────────────────────────────────────
	contentsLines := []string{
		"",
		"  MSX BASIC HELP CONTENTS",
		"",
		"  How to Use Help",
		"  Menus and Hot Keys",
		"  Editor Commands",
		"",
		"  ─────────────────────────────────────────────────────",
		"",
		"  Built-in",
		"  Command Line",
		"  Debugging",
		"  Directives",
		"  Error Messages",
		"  Browser",
		"",
		"  ─────────────────────────────────────────────────────",
		"",
		"  Reserved Worlds",
		"  Start-Up Options",
		"  MSXgl",
		"  Library",
	}
	contentsLinks := buildLinks(contentsLines, []linkDef{
		{"How to Use Help", "how_to_use"},
		{"Menus and Hot Keys", "menus_hotkeys"},
		{"Editor Commands", "editor_commands"},
		{"Built-in", "built_in"},
		{"Command Line", "command_line"},
		{"Debugging", "debugging"},
		{"Directives", "directives"},
		{"Error Messages", "error_messages"},
		{"Browser", "browser"},
		{"Reserved Worlds", "reserved_worlds"},
		{"Start-Up Options", "startup_options"},
		{"MSXgl", "msxgl"},
		{"Library", "library"},
	})
	hc.topics["contents"] = &HelpTopic{
		ID:    "contents",
		Title: "Help",
		Lines: contentsLines,
		Links: contentsLinks,
	}

	// ── Tópico: How to Use Help ───────────────────────────────────────────────
	hc.topics["how_to_use"] = hc.makePlaceholder("how_to_use", "How to Use Help",
		[]string{
			"",
			"  HOW TO USE HELP",
			"",
			"  Use Tab / Shift+Tab to move between links.",
			"  Press Enter to follow a link.",
			"  Press Alt+F1 to return to the previous topic.",
			"  Press Escape to close Help.",
			"",
			"  Contents",
		},
		[]linkDef{{"Contents", "contents"}},
	)

	// ── Tópico: Menus and Hot Keys ────────────────────────────────────────────
	hc.topics["menus_hotkeys"] = hc.makePlaceholder("menus_hotkeys", "Menus and Hot Keys",
		[]string{
			"",
			"  MENUS AND HOT KEYS",
			"",
			"  F1          Open Help",
			"  F2          Save",
			"  F3          Open",
			"  Alt+F1      Previous Help topic",
			"  Alt+X       Exit",
			"  F10         File menu",
			"",
			"  Contents",
		},
		[]linkDef{{"Contents", "contents"}},
	)

	// ── Tópico: Editor Commands ───────────────────────────────────────────────
	hc.topics["editor_commands"] = hc.makePlaceholder("editor_commands", "Editor Commands",
		[]string{
			"",
			"",
			"",
			"      Block commands",
			"      Cursor-movement commands",
			"      Insert & Delete commands",
			"      Miscelaneous commands",
			"      Syntax highlighting",
			"",
			"      Contents",
		},
		[]linkDef{
			{"Block commands", "block_commands"},
			{"Cursor-movement commands", "cursor_movement_commands"},
			{"Insert & Delete commands", "insert_delete_commands"},
			{"Miscelaneous commands", "miscelaneous_commands"},
			{"Syntax highlighting", "syntax_highlighting"},
			{"Contents", "contents"},
		},
	)

	// ── Tópico: Block Commands ────────────────────────────────────────────────
	blockCmdLines := []string{
		"",
		"  Block Commands", // linha 1: renderizada como botão 3D em help_window.go
		"",
		"  Ctrl+K B    Mark block begin",
		"  Ctrl+K K    Mark block end",
		"  Ctrl+K T    Mark single word",
		"  Ctrl+K L    Mark line",
		"  Ctrl+K C    Copy block",
		"  Ctrl+K V    Move block",
		"  Ctrl+K Y    Delete block",
		"  Ctrl+K R    Read block from disk",
		"  Ctrl+K W    Write block to disk",
		"  Ctrl+K H    Hide/display block",
		"  Ctrl+K P    Print block",
		"  Ctrl+K I    Indent block",
		"  Ctrl+K U    Unindent block",
		"  Ctrl+K D    Exit to menu bar",
		"  Ctrl+Q B    Move to begin of block",
		"  Ctrl+Q K    Move to end of block",
		"  Ctrl+Ins    Copy to clipboard",
		"  Shift+Del   Cut to clipboard",
		"  Ctrl+Del    Delete block",
		"  Shift+Ins   Paste from clipboard",
		"",
		"  See Also:",
		"    Extending Selected Blocks",
		"",
		"  Contents",
	}
	hc.topics["block_commands"] = &HelpTopic{
		ID:    "block_commands",
		Title: "Block commands",
		Lines: blockCmdLines,
		Links: buildLinks(blockCmdLines, []linkDef{
			{"Extending Selected Blocks", "extending_selected_blocks"},
			{"Contents", "contents"},
		}),
	}

	// ── Tópicos placeholder ───────────────────────────────────────────────────
	type simpleTopic struct {
		id    string
		title string
	}
	simple := []simpleTopic{
		{"built_in", "Built-in"},
		{"command_line", "Command Line"},
		{"cursor_movement_commands", "Cursor-movement commands"},
		{"debugging", "Debugging"},
		{"directives", "Directives"},
		{"error_messages", "Error Messages"},
		{"extending_selected_blocks", "Extending Selected Blocks"},
		{"insert_delete_commands", "Insert & Delete commands"},
		{"browser", "Browser"},
		{"miscelaneous_commands", "Miscelaneous commands"},
		{"reserved_worlds", "Reserved Worlds"},
		{"startup_options", "Start-Up Options"},
		{"syntax_highlighting", "Syntax highlighting"},
		{"msxgl", "MSXgl"},
		{"library", "Library"},
	}
	for _, s := range simple {
		id, title := s.id, s.title
		hc.topics[id] = hc.makePlaceholder(id, title,
			[]string{
				"",
				"  " + strings.ToUpper(title),
				"",
				"  [Content to be implemented]",
				"",
				"  Contents",
			},
			[]linkDef{{"Contents", "contents"}},
		)
	}
}

// makePlaceholder cria um HelpTopic com posições de links calculadas automaticamente
func (hc *HelpContent) makePlaceholder(id, title string, lines []string, defs []linkDef) *HelpTopic {
	return &HelpTopic{
		ID:    id,
		Title: title,
		Lines: lines,
		Links: buildLinks(lines, defs),
	}
}

// GetCurrentTopic retorna o tópico atual
func (hc *HelpContent) GetCurrentTopic() *HelpTopic {
	if topic, ok := hc.topics[hc.currentTopicID]; ok {
		return topic
	}
	return hc.topics["contents"]
}

// NavigateToTopic navega para um novo tópico
func (hc *HelpContent) NavigateToTopic(topicID string) bool {
	if _, ok := hc.topics[topicID]; !ok {
		return false
	}
	hc.history = append(hc.history, hc.currentTopicID)
	hc.currentTopicID = topicID
	return true
}

// GoBack volta ao tópico anterior
func (hc *HelpContent) GoBack() bool {
	if len(hc.history) == 0 {
		return false
	}
	prevTopicID := hc.history[len(hc.history)-1]
	hc.history = hc.history[:len(hc.history)-1]
	hc.currentTopicID = prevTopicID
	return true
}

// BreadcrumbTrail retorna o caminho virtual atual, do topo raiz ao tópico atual.
func (hc *HelpContent) BreadcrumbTrail() []string {
	trail := make([]string, 0, len(hc.history)+1)
	for _, id := range hc.history {
		if topic, ok := hc.topics[id]; ok {
			trail = append(trail, topic.Title)
		} else {
			trail = append(trail, id)
		}
	}
	if topic, ok := hc.topics[hc.currentTopicID]; ok {
		trail = append(trail, topic.Title)
	} else {
		trail = append(trail, hc.currentTopicID)
	}
	return trail
}

// BreadcrumbText retorna o breadcrumb formatado para exibição.
func (hc *HelpContent) BreadcrumbText() string {
	trail := hc.BreadcrumbTrail()
	return strings.Join(trail, " > ")
}

// FindNextLink encontra o próximo link a partir de currentLinkIdx
func (hc *HelpContent) FindNextLink(currentLinkIdx int, forward bool) int {
	topic := hc.GetCurrentTopic()
	if len(topic.Links) == 0 {
		return -1
	}
	if currentLinkIdx < 0 {
		if forward {
			return 0
		}
		return len(topic.Links) - 1
	}
	if forward {
		return (currentLinkIdx + 1) % len(topic.Links)
	}
	return (currentLinkIdx - 1 + len(topic.Links)) % len(topic.Links)
}

// GetLinkAtIndex retorna o link no índice especificado
func (hc *HelpContent) GetLinkAtIndex(idx int) *LinkInfo {
	topic := hc.GetCurrentTopic()
	if idx >= 0 && idx < len(topic.Links) {
		return &topic.Links[idx]
	}
	return nil
}

// GetPagedContent retorna as linhas visíveis a partir de startRow
func (hc *HelpContent) GetPagedContent(startRow int, maxRows int) []string {
	topic := hc.GetCurrentTopic()
	if startRow < 0 {
		startRow = 0
	}
	endRow := startRow + maxRows
	if endRow > len(topic.Lines) {
		endRow = len(topic.Lines)
	}
	if startRow >= len(topic.Lines) {
		return []string{}
	}
	return topic.Lines[startRow:endRow]
}

// GetMaxLine retorna o número total de linhas do tópico atual
func (hc *HelpContent) GetMaxLine() int {
	return len(hc.GetCurrentTopic().Lines)
}

// GetMaxCol retorna a largura da linha mais longa do tópico atual
func (hc *HelpContent) GetMaxCol() int {
	maxCol := 0
	for _, line := range hc.GetCurrentTopic().Lines {
		if l := len([]rune(line)); l > maxCol {
			maxCol = l
		}
	}
	return maxCol
}

