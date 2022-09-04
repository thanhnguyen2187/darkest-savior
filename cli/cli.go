package cli

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"darkest-savior/dson"
	"darkest-savior/dson/dheader"
	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

type (
	Args struct {
		Interactive *InteractiveCmd `arg:"subcommand:interactive"`
		Convert     *ConvertCmd     `arg:"subcommand:convert"`
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
	print("Not implemented!")
}

func CheckExistence(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return err == nil
}

func StartConverting(from string, to string, force bool) {
	if !CheckExistence(from) {
		println("Source file does not exist!")
		return
	}
	if CheckExistence(to) && !force {
		println("Destination file existed. Please type the command again with --force to allow overwriting!")
		println("Explicit --force is needed to make sure that you paid attention not to overwriting the actual DSON file in your folder.")
		return
	}
	fileBytes, err := ioutil.ReadFile(from)
	if err != nil {
		println("Error happened reading file")
		return
	}

	if dheader.IsValidMagicNumber(fileBytes[:4]) {
		// TODO: Improve the interfaces by letting `cli` knows nothing about what is `DecodedFile`.
		//       Let the data exchange format between `cli` and `dson` be `[]byte`
		decodedFile, err := dson.DecodeDSON(fileBytes)
		if err != nil {
			println("Error happened decoding DSON to JSON")
			return
		}
		decodedMap := dson.ToLinkedHashMap(*decodedFile)
		decodedBytes, err := json.MarshalIndent(decodedMap, "", "  ")
		if err := ioutil.WriteFile(to, decodedBytes, 0644); err != nil {
			println("Error happened writing to file at: " + to)
			return
		}
		println("Done converting. Please check your result file at: " + to)
	} else {
		bs, err := ioutil.ReadFile(from)
		rbs, err := dson.EncodeDSON(bs)
		if err != nil {
			println("Error happened encoding JSON to DSON")
			return
		}
		err = ioutil.WriteFile(to, rbs, 0644)
		if err != nil {
			println("Error happened writing output to: " + to)
			return
		}
	}
}

func Start() {
	args := Args{}
	arg.MustParse(&args)

	if (args.Interactive == nil && args.Convert == nil) ||
		args.Interactive != nil {
		StartInteractive()
	} else {
		StartConverting(
			args.Convert.From,
			args.Convert.To,
			args.Convert.Force,
		)
	}
}
