package cli

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
	"github.com/thanhnguyen2187/darkest-savior/dson"
)

type (
	Args struct {
		// Interactive *InteractiveCmd `arg:"subcommand:interactive"`
		Convert *ConvertCmd `arg:"subcommand:convert"`
	}
	InteractiveCmd struct{}
	ConvertCmd     struct {
		// TODO: improve UX of `from` and `to`
		// the underlying library has some limitation on displaying help and placeholder
		// too long placeholder force help to be put on another line, which looks really ugly
		// that is why text is really sparse for the arguments, even though I wanted it to be clearer
		From  string `arg:"required" help:"path to source file" placeholder:"persist.json"`
		To    string `arg:"required" help:"path to destination file" placeholder:"file.json"`
		Force bool   `help:"overwrite the destination file"`
		Debug bool   `help:"enable debugging on destination file"`
	}
)

func (Args) Description() string {
	des := strings.Join(
		[]string{
			"Ruin has come to our command line.\n",
			"A CLI utility to convert DSON (Darkest Dungeon's own proprietary JSON format)",
			"to \"standard\" JSON in the command line.",
		},
		"\n",
	)
	des += "\n"
	return des
}

func StartInteractive() {
	println("Not implemented! Please just use  convert  for now.")
}

func CheckExistence(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return err == nil
}

func StartConverting(args ConvertCmd) {
	if !CheckExistence(args.From) {
		println("Source file does not exist!")
		return
	}
	if CheckExistence(args.To) && !args.Force {
		println("Destination file existed. Please type the command again with --force to allow overwriting!")
		println("Explicit --force is needed to make sure that you paid attention not to overwriting the actual DSON file in your folder.")
		return
	}
	fileBytes, err := ioutil.ReadFile(args.From)
	if err != nil {
		println("Error happened reading file")
		return
	}

	if dson.IsDSONFile(fileBytes[:4]) {
		resultBytes, err := dson.DecodeDSON(fileBytes, args.Debug)
		if err != nil {
			println("Error happened decoding DSON to JSON")
			return
		}
		if err := ioutil.WriteFile(args.To, resultBytes, 0644); err != nil {
			println("Error happened writing to file at: " + args.To)
			return
		}
	} else {
		resultBytes, err := dson.EncodeJSON(fileBytes)
		if err != nil {
			println("Error happened encoding JSON to DSON")
			return
		}
		err = ioutil.WriteFile(args.To, resultBytes, 0644)
		if err != nil {
			println("Error happened writing output to: " + args.To)
			return
		}
	}
	println("Done converting. Please check your result file at: " + args.To)
}

func Start() {
	args := Args{}
	arg.MustParse(&args)

	if args.Convert != nil {
		StartConverting(*args.Convert)
	} else {
		println("Convert from DSON to JSON and vice versa are available.")
		println("Please use the functionality by retyping your command with `convert` at the end.")
	}
}
