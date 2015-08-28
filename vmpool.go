package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

var version string

func init() {
	//  vlcoud base configuration can be overidden via ENV variable
	if vmpool_url = os.Getenv("VMPOOL_URL"); vmpool_url == "" {
		vmpool_url = "https://vmpooler.delivery.puppetlabs.net"
	}
}

var usageTemplate = `Vmpool is a tool for retrieving remote vms.

Usage:

     vmpool command [arguments]

The commands are:
{{range .}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "vmpool help [command]" for more information about a command.

`

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Usage prints the
func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Name returns the first word of the UsageLine
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

var commands = []*Command{
	cmdGrab,
	cmdDelete,
	cmdList,
	cmdStatus,
	cmdSummary,
	cmdToken,
	cmdVersion,
	cmdVm,
}

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("foo")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	err := t.Execute(w, data)
	perror(err)
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

var vmpool_url string

var helpTemplate = `usage: vmpool {{.UsageLine}}

{{.Long | trim}}
`

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'vmpool help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: vmpool help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'vmpool help'
	}

	arg := args[0]

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'vmpool help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'vmpool help'.\n", arg)
	os.Exit(2) // failed at 'vmpool help cmd'
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			cmd.Run(cmd, args)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "vmpool: unknown subcommand %q\nRun 'vmpool help' for usage.\n", args[0])
	os.Exit(2)
}
