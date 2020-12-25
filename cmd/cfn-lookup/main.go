package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fujiwara/cfn-lookup/cfn"
	"github.com/pkg/errors"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var cache sync.Map

func _main() error {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		return flag.ErrHelp
	}

	subcmd := args[0]
	app := cfn.New(session.Must(session.NewSession()), &cache)
	switch subcmd {
	case "output":
		if len(args) != 3 {
			return errors.Errorf("required outputKey and outputValue by arguments")
		}
		return cmdOutput(app, args[1:])
	case "export":
		if len(args) == 1 {
			return cmdExportList(app)
		}
		return cmdExport(app, args[1:])
	}
	return errors.Errorf("invalid command %s", subcmd)
}

func cmdOutput(app *cfn.App, args []string) error {
	value, err := app.LookupOutput(args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

func cmdExportList(app *cfn.App) error {
	names, err := app.ExportedNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func cmdExport(app *cfn.App, args []string) error {
	for _, name := range args {
		value, err := app.LookupExport(name)
		if err != nil {
			return err
		}
		fmt.Println(value)
	}
	return nil
}
