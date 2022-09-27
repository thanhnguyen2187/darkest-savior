package ui

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/thanhnguyen2187/darkest-savior/match"
)

const (
	CwdStateCorrect   = "correct"
	CwdStateIncorrect = "incorrect"
	CwdStateBlank     = ""
)

type FileSelector struct {
	cwd      string
	cwdState string
}

func CreateFileSelector() FileSelector {
	cwd, err := os.Getwd()
	if err != nil {
		err := errors.Wrap(err, "CreateFileSelector get current working directory error")
		log.Panic(err)
	}
	return FileSelector{
		cwd:      cwd,
		cwdState: "",
	}
}

type FileName string

func ReadDirectory(path string) []FileName {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	fileNames := lo.Map(
		files,
		func(t fs.FileInfo, _ int) FileName {
			return FileName(t.Name())
		},
	)
	return fileNames
}

func (s FileSelector) View() string {
	output := "DARKEST SAVIOR\n\n"
	output += "Current directory: " + s.cwd + "\n"

	_, msgAny := match.
		Match(s.cwdState).
		When(
			match.OneOf(CwdStateIncorrect, CwdStateBlank),
			func() string {
				return "Please choose the correct save folder"
			},
		).
		When(
			CwdStateCorrect,
			func() string {
				return "Looks like a valid save folder"
			},
		).
		When(
			match.ANY,
			func() string {
				err := fmt.Sprintf(`FileSelector.View unreachable code: invalid state of current directory "%s"`, s.cwdState)
				log.Panic(err)
				return ""
			},
		).
		Result()
	msg := msgAny.(string)
	output += msg

	return output
}

func (s FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

func (s FileSelector) Init() tea.Cmd {
	return nil
}
