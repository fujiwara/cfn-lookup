package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fujiwara/cfn-lookup/cfn"
	"github.com/google/subcommands"
)

var cache sync.Map

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&outputCmd{}, "")
	subcommands.Register(&exportCmd{}, "")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	app := cfn.New(cfg, &cache)
	os.Exit(int(subcommands.Execute(ctx, app)))
}

type outputCmd struct {
	list bool
}

func (*outputCmd) Name() string     { return "output" }
func (*outputCmd) Synopsis() string { return "Lookup an output value from the stack" }
func (c *outputCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.list, "list", false, "show output keys")
}
func (c *outputCmd) Usage() string {
	return `output [-list] StackName [OutputKey]:
Lookup an OutputValue of the OutputKey in the StackName.`
}

func (c *outputCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	app := args[0].(*cfn.App)
	ag := f.Args()

	if c.list {
		if len(ag) != 1 {
			fmt.Fprintln(os.Stderr, c.Usage())
			return subcommands.ExitFailure
		}
		keys, err := app.ListOutput(ctx, ag[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return subcommands.ExitFailure
		}
		fmt.Println(strings.Join(keys, "\n"))
	} else {
		if len(ag) != 2 {
			fmt.Fprintln(os.Stderr, c.Usage())
			return subcommands.ExitFailure
		}
		value, err := app.LookupOutput(ctx, ag[0], ag[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return subcommands.ExitFailure
		}
		fmt.Println(value)
	}
	return subcommands.ExitSuccess

}

type exportCmd struct {
	list bool
}

func (*exportCmd) Name() string     { return "export" }
func (*exportCmd) Synopsis() string { return "Lookup an exported value from CFn exports" }
func (c *exportCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.list, "list", false, "show exported names")
}
func (*exportCmd) Usage() string {
	return `export [-list] Name:
Lookup an exported value of the Name.`
}

func (c *exportCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	app := args[0].(*cfn.App)

	if c.list {
		names, err := app.ExportedNames(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return subcommands.ExitFailure
		}
		fmt.Println(strings.Join(names, "\n"))
	} else {
		for _, name := range f.Args() {
			value, err := app.LookupExport(ctx, name)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return subcommands.ExitFailure
			}
			fmt.Println(value)
		}
	}

	return subcommands.ExitSuccess
}
