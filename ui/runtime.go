package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Start() {
	folderSelector := CreateFileSelector()
	if err := tea.NewProgram(&folderSelector).Start(); err != nil {
		panic(err)
	}
}
