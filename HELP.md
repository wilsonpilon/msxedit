# Help Contents

MSXEDIT HELP CONTENTS

Use `Tab` or `Shift+Tab` to move between links, `Enter` to open a topic, `Alt+F1` to go back, and `Esc` to close Help.

- [How to Use Help](#how-to-use-help)
- [Menus and Hot Keys](#menus-and-hot-keys)
- [Editor Commands](#editor-commands)
- [Built-in](#built-in)
- [Command Line](#command-line)
- [Debugging](#debugging)
- [Directives](#directives)
- [Error Messages](#error-messages)
- [Browser](#browser)
- [Reserved Words](#reserved-words)
- [Start-Up Options](#start-up-options)
- [MSXgl](#msxgl)
- [Library](#library)

## How to Use Help

The Help window in MSXEdit is a full custom TUI window with its own frame, scrollbars, topic title and window number.

Keyboard navigation:

- `Tab` / `Shift+Tab`: select next or previous link
- `Enter`: open selected link
- `Alt+F1`: return to previous topic
- `Alt+Q`: fallback back-navigation for terminals that do not forward `Alt+F1`
- `PgUp`, `PgDn`, `Home`, `End`: fast scrolling
- `Esc`: close Help

Mouse navigation:

- click a link to open it
- click `[■]` to close the Help window
- click the vertical scrollbar to scroll by position
- click the bottom scrollbar area to move horizontally

References:
- [Help Contents](#contents)
- [Menus and Hot Keys](#menus-and-hot-keys)

## Menus and Hot Keys

The top menu bar follows the classic Turbo style and can be opened by keyboard or mouse.

Implemented hot keys:

- `F1`: open Help
- `F10`: open the `File` menu
- `Alt+X`: exit MSXEdit
- `Alt+F`, `Alt+E`, `Alt+S`, `Alt+R`, `Alt+C`, `Alt+D`, `Alt+T`, `Alt+O`, `Alt+W`, `Alt+H`: open the matching menu

Current menu status:

- `File`: active menu, with `Exit` implemented
- `Help`: active menu, with `Contents` and `About` implemented
- `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Tools`, `Options`, `Window`: visible and navigable, still scaffolded with placeholder content

The status bar also displays legacy labels for `F2 Save`, `F3 Open`, `Alt+F9 Compile` and `F9 Make`, but those workflows are not yet wired to concrete actions.

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)
- [Command Line](#command-line)

## Editor Commands

This topic groups the current editing-related areas.

- [Block commands](#block-commands)
- [Cursor-movement commands](#cursor-movement-commands)
- [Insert & Delete commands](#insert-delete-commands)
- [Miscelaneous commands](#miscelaneous-commands)
- [Syntax highlighting](#syntax-highlighting)

References:
- [Help Contents](#contents)
- [How to Use Help](#how-to-use-help)

## Built-in

Built-in components already used by the application:

- custom editor window with double border
- `dialogoOK` reusable dialog
- `turboButton` reusable Turbo-style button
- `About` dialog
- `Help` window with markdown topic loading and fallback internal content

References:
- [Help Contents](#contents)
- [Library](#library)

## Command Line

MSXEdit currently supports the following command line options:

- `--theme <default|blue>`
- `--tabsize <n>`
- `--no-highlight`
- `--local`
- `--version`

You can also pass an optional file name as the last argument. When present, that path is used as the title of the first editor window.

References:
- [Help Contents](#contents)
- [Start-Up Options](#start-up-options)

## Debugging

The current focus of the project is UI structure and navigation. Debugging tools such as source-level stepping, watches and runtime inspectors are not yet available from the `Debug` menu.

At this stage, the most visible diagnostics are:

- CLI usage and version output
- config loading warnings in terminal mode
- markdown Help fallback when `HELP.md` cannot be read

References:
- [Help Contents](#contents)
- [Error Messages](#error-messages)

## Directives

Compiler or project directives are not yet interpreted by the editor UI. This topic is reserved for future language-aware behavior.

At the moment, the nearest related pieces are the planned syntax/highlight pipeline and the language-specific roadmap tracked in the project documentation.

References:
- [Help Contents](#contents)
- [Syntax highlighting](#syntax-highlighting)

## Error Messages

Known runtime/user-visible situations include:

- configuration file cannot be decoded
- configuration file does not exist and defaults are used
- `HELP.md` cannot be found, so built-in topics are loaded instead
- tokenized BASIC loader is still not implemented

References:
- [Help Contents](#contents)
- [Debugging](#debugging)

## Browser

There is no full file browser dialog yet.

What already exists:

- editor startup with optional filename argument
- menu scaffolding for future file actions
- mouse-aware top menu interaction

What is still pending:

- `Open...`
- directory changes from the UI
- save dialogs and file browsing panels

References:
- [Help Contents](#contents)
- [Menus and Hot Keys](#menus-and-hot-keys)

## Reserved Words

Language keyword databases are still in progress.

The repository already contains an initial MSX-BASIC token table that will support future language-aware features, but no keyword browser is exposed in the UI yet.

References:
- [Help Contents](#contents)
- [MSXgl](#msxgl)

## Start-Up Options

Startup behavior summary:

- no argument: open editor window `1` with title `Sem Nome`
- file argument: open editor window `1` using the provided path as title
- `--theme`: choose the VGA palette profile
- `--local`: load `msxedit.json` from the current directory

The application always opens the main editor window automatically.

References:
- [Help Contents](#contents)
- [Command Line](#command-line)

## MSXgl

MSXgl remains part of the long-term language support roadmap. The current release does not yet provide dedicated editing assistance, compile integration or contextual help for MSXgl symbols.

References:
- [Help Contents](#contents)
- [Library](#library)

## Library

Current reusable UI building blocks include:

- `dialogoOK`
- `turboButton`
- `helpHeaderButton`
- theme palette helpers
- markdown Help loader and parser

These pieces are intended to keep future dialogs and tool windows visually consistent.

References:
- [Help Contents](#contents)
- [Built-in](#built-in)

## Block commands

Dedicated block selection, cut, copy and paste workflows are not exposed yet as custom editor commands.

The current editor area is a text entry surface with retro window chrome. Advanced block operations remain a future step.

References:
- [Editor Commands](#editor-commands)
- [Help Contents](#contents)

## Cursor-movement commands

The editor and Help window already support cursor/scroll movement through standard keyboard navigation.

In Help specifically, arrow keys, page navigation and Home/End are implemented.

References:
- [Editor Commands](#editor-commands)
- [Help Contents](#contents)

## Insert & Delete commands

Text insertion and deletion are currently handled by the underlying text area component used by the editor window.

Higher-level editing commands and retro key maps are still candidates for future implementation.

References:
- [Editor Commands](#editor-commands)
- [Help Contents](#contents)

## Miscelaneous commands

Other current interaction points include:

- `About` dialog
- menu navigation
- window closing by `[■]`
- mouse support in Help and menus

References:
- [Editor Commands](#editor-commands)
- [Help Contents](#contents)

## Syntax highlighting

The project already exposes the configuration flag `highlight` and the CLI switch `--no-highlight`, but the actual syntax coloring pipeline is not fully connected to the editor surface yet.

This topic is therefore a status note rather than a completed feature description.

References:
- [Editor Commands](#editor-commands)
- [Directives](#directives)

