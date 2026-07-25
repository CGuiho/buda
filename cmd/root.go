package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/agent"
	clihelp "github.com/CGuiho/buda/internal/help"
	"github.com/CGuiho/buda/internal/maintenance"
	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Target  string `json:"target"`
}

// Options contains the stable persistent options shared by every command.
// Command constructors receive the same pointer through Dependencies.
type Options struct {
	Wiki string
	JSON bool
}

type Dependencies struct {
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Options *Options
	Agents  *agent.Service

	Executable          func() (string, error)
	ScheduleMaintenance func(executable, wiki string) error
}

type exitCoder interface {
	ExitCode() int
}

type codedError struct {
	code    int
	message string
	cause   error
}

func (e *codedError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

func (e *codedError) Unwrap() error { return e.cause }
func (e *codedError) ExitCode() int { return e.code }

var errHelpRendered = errors.New("developer context help rendered")

func UsageError(format string, values ...any) error {
	return &codedError{code: 2, message: fmt.Sprintf(format, values...)}
}

func RepositoryError(message string, cause error) error {
	return &codedError{code: 3, message: message, cause: cause}
}

func MutationError(message string, cause error) error {
	return &codedError{code: 5, message: message, cause: cause}
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		In:                  os.Stdin,
		Out:                 os.Stdout,
		Err:                 os.Stderr,
		Options:             &Options{},
		Agents:              agent.NewService(agent.DefaultResources()),
		Executable:          os.Executable,
		ScheduleMaintenance: maintenance.Schedule,
	}
}

func Execute(info BuildInfo) error {
	deps := DefaultDependencies()
	root := NewRootCommand(deps, info, NewApplicationCommands(deps)...)
	err := root.Execute()
	if errors.Is(err, errHelpRendered) {
		return nil
	}
	return err
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var coded exitCoder
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	return 1
}

// NewRootCommand returns a fresh Cobra tree. Public repository commands are
// injected explicitly; no package-global command registry is used.
func NewRootCommand(deps Dependencies, info BuildInfo, commands ...*cobra.Command) *cobra.Command {
	deps = normalizeDependencies(deps)
	options := deps.Options
	var helpTree, helpDocs bool
	var helpDepth int

	root := &cobra.Command{
		Use:           "buda",
		Short:         "Maintain and retrieve one explicit evidence-backed OKF wiki.",
		Version:       info.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          NoArgs,
		PersistentPreRunE: func(command *cobra.Command, _ []string) error {
			if command.Flags().Changed("help-tree-depth") && helpDepth < 1 {
				return UsageError("--help-tree-depth must be a positive integer")
			}
			if helpDocs {
				markdown, err := clihelp.Markdown(command, helpDepth)
				if err != nil {
					return err
				}
				fmt.Fprint(command.OutOrStdout(), markdown)
				return errHelpRendered
			}
			if helpTree || command.Flags().Changed("help-tree-depth") {
				fmt.Fprint(command.OutOrStdout(), clihelp.Tree(command, helpDepth))
				return errHelpRendered
			}
			if isRepositoryCommand(command) {
				if strings.TrimSpace(options.Wiki) == "" {
					return RepositoryError("--wiki is required; Buda never selects a wiki implicitly", nil)
				}
				absolute, err := filepath.Abs(options.Wiki)
				if err != nil {
					return RepositoryError("resolve --wiki", err)
				}
				options.Wiki = filepath.Clean(absolute)
			}
			return nil
		},
		PersistentPostRunE: func(command *cobra.Command, _ []string) error {
			if !maintenance.ShouldSchedule(command) || options.Wiki == "" {
				return nil
			}
			executable, err := deps.Executable()
			if err != nil {
				return nil
			}
			_ = deps.ScheduleMaintenance(executable, options.Wiki)
			return nil
		},
		RunE: func(command *cobra.Command, _ []string) error {
			message := fmt.Sprintf("Hello Windows - buda v%s", info.Version)
			if options.JSON {
				return WriteJSON(command, map[string]any{
					"command": "buda", "version": info.Version, "message": message,
				})
			}
			fmt.Fprintln(command.OutOrStdout(), message)
			return nil
		},
	}
	root.SetIn(deps.In)
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)
	root.SetVersionTemplate("{{.Name}} v{{.Version}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Use: "help", Hidden: true})
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return UsageError("parse flags: %v", err)
	})

	flags := root.PersistentFlags()
	flags.StringVar(&options.Wiki, "wiki", "", "Explicit path to the one wiki repository")
	flags.BoolVar(&options.JSON, "json", false, "Emit one stable JSON document")
	flags.BoolVar(&helpTree, "help-tree", false, "Show the public command subtree")
	flags.IntVar(&helpDepth, "help-tree-depth", 0, "Limit command-tree recursion to a positive depth")
	flags.BoolVar(&helpDocs, "help-docs", false, "Show deterministic Markdown for this command scope")

	root.AddCommand(commands...)
	root.AddCommand(newAgentCommand(deps))
	root.AddCommand(newMaintenanceCommand(deps))
	return root
}

func normalizeDependencies(deps Dependencies) Dependencies {
	if deps.In == nil {
		deps.In = strings.NewReader("")
	}
	if deps.Out == nil {
		deps.Out = io.Discard
	}
	if deps.Err == nil {
		deps.Err = io.Discard
	}
	if deps.Options == nil {
		deps.Options = &Options{}
	}
	if deps.Agents == nil {
		deps.Agents = agent.NewService(agent.DefaultResources())
	}
	if deps.Executable == nil {
		deps.Executable = os.Executable
	}
	if deps.ScheduleMaintenance == nil {
		deps.ScheduleMaintenance = maintenance.Schedule
	}
	return deps
}

func isRepositoryCommand(command *cobra.Command) bool {
	if command == command.Root() {
		return false
	}
	top := command
	for top.Parent() != nil && top.Parent() != command.Root() {
		top = top.Parent()
	}
	return top.Name() != "agent" && !top.Hidden
}

func NoArgs(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return UsageError("accepts 0 arg(s), received %d", len(args))
	}
	return nil
}

func ExactArgs(count int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != count {
			return UsageError("accepts %d arg(s), received %d", count, len(args))
		}
		return nil
	}
}

// WriteJSON writes exactly one indented JSON document followed by one newline.
func WriteJSON(command *cobra.Command, value any) error {
	encoder := json.NewEncoder(command.OutOrStdout())
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode JSON output: %w", err)
	}
	return nil
}

func JSONRequested(deps Dependencies) bool {
	return deps.Options != nil && deps.Options.JSON
}
