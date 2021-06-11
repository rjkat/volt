package subcmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/rjkat/volt/subcmd/builder"
	"github.com/rjkat/volt/transaction"
)

func init() {
	cmdMap["build"] = &buildCmd{}
}

type buildCmd struct {
	helped bool
	full   bool
}

func (cmd *buildCmd) ProhibitRootExecution(args []string) bool { return true }

func (cmd *buildCmd) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  volt build [-help] [-full]

Quick example
  $ volt build        # builds directories under ~/.vim/pack/volt
  $ volt build -full  # full build (remove ~/.vim/pack/volt, and re-create all)

Description
  Build ~/.vim/pack/volt/opt/ directory:
    1. Copy repositories' files into ~/.vim/pack/volt/opt/
      * If the repository is git repository, extract files from locked revision of tree object and copy them into above vim directories
      * If the repository is static repository (imported non-git directory by "volt add" command), copy files into above vim directories
    2. Remove directories from above vim directories, which exist in ~/.vim/pack/volt/build-info.json but not in $VOLTPATH/lock.json

  ~/.vim/pack/volt/build-info.json is a file which holds the information that what vim plugins are installed in ~/.vim/pack/volt/ and its type (git repository, static repository, or system repository), its version. A user normally doesn't need to know the contents of build-info.json .

  If -full option was given, remove all directories in ~/.vim/pack/volt/opt/ , and copy repositories' files into above vim directories.
  Otherwise, it will perform smart build: copy / remove only changed repositories' files.` + "\n\n")
		fmt.Println("Options")
		fs.PrintDefaults()
		fmt.Println()
		cmd.helped = true
	}
	fs.BoolVar(&cmd.full, "full", false, "full build")
	return fs
}

func (cmd *buildCmd) Run(args []string) (result *Error) {
	// Parse args
	fs := cmd.FlagSet()
	fs.Parse(args)
	if cmd.helped {
		return nil
	}

	// Begin transaction
	trx, err := transaction.Start()
	if err != nil {
		result = &Error{Code: 11, Msg: "Failed to begin transaction: " + err.Error()}
		return
	}
	defer func() {
		if err := trx.Done(); err != nil {
			result = &Error{Code: 13, Msg: "Failed to end transaction: " + err.Error()}
		}
	}()

	err = builder.Build(cmd.full)
	if err != nil {
		result = &Error{Code: 12, Msg: "Failed to build: " + err.Error()}
		return
	}

	return
}
