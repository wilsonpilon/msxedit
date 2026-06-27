package tui

import (
	"fmt"
)

type aboutDialog struct {
	*dialogoOK
}

func newAboutDialog(app *App) *aboutDialog {
	text := []string{
		"MSXEdit",
		"",
		fmt.Sprintf("Versao %s", app.Version),
		"",
		"Copyright (c) 2026 by",
		"Junie AI Editor",
	}

	dlg := newDialogoOK(app, "about", text, nil)
	dlg.SetButton("O&K", 'k', nil)
	dlg.SetButtonShadowMode(shadowModeTurboClassic)

	return &aboutDialog{dialogoOK: dlg}
}
