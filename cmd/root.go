package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CGuiho/buda/examples"
	"github.com/CGuiho/buda/internal/agent"
	clihelp "github.com/CGuiho/buda/internal/help"
	"github.com/CGuiho/buda/internal/installlayout"
	"github.com/CGuiho/buda/internal/maintenance"
	"github.com/CGuiho/buda/internal/qmd"
	"github.com/CGuiho/buda/internal/releasecatalog"
	"github.com/CGuiho/buda/internal/repository"
	"github.com/CGuiho/buda/internal/selfmanage"
	"github.com/CGuiho/buda/internal/source"
	"github.com/CGuiho/buda/internal/upgrade"
	"github.com/CGuiho/buda/prompts"
	"github.com/CGuiho/buda/schemas"
	"github.com/CGuiho/buda/skills"
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
	In  io.Reader
	Out io.Writer
	Err io.Writer
	// Version is the exact release version used for generated, version-pinned
	// configuration metadata. Development/test callers may leave it empty and
	// receive the repository's current fallback.
	Version       string
	HomeDir       func() (string, error)
	InstallLayout func() (installlayout.Layout, error)
	// Interactive reports whether prompts may be shown. Keeping terminal
	// detection injectable lets init remain deterministic in tests and fail
	// closed when an unattended invocation lacks required answers.
	Interactive func() bool
	Options     *Options
	Agents      *agent.Service

	Executable          func() (string, error)
	ScheduleMaintenance func(executable, wiki string) error
	Now                 func() time.Time
	HTTPClient          source.Doer
	RemoveExecutable    func(string) (bool, error)
	RollbackExecutable  func(string) (bool, error)
	UpgradeRelease      func(context.Context, upgrade.Options) (upgrade.Result, error)
	ReconcileInstalled  func(string, string) error
}

type exitCoder interface {
	ExitCode() int
}

type codedError struct {
	code    int
	message string
	cause   error
}

type renderedError struct{ cause error }

func (e *renderedError) Error() string { return e.cause.Error() }
func (e *renderedError) Unwrap() error { return e.cause }
func (e *renderedError) ExitCode() int { return ExitCode(e.cause) }

type errorDocument struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	QMD     *qmd.CommandError `json:"qmd,omitempty"`
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
		Interactive:         func() bool { return interactiveReader(os.Stdin) },
		Options:             &Options{},
		Agents:              agent.NewService(agent.DefaultResources()),
		Executable:          os.Executable,
		HomeDir:             os.UserHomeDir,
		InstallLayout:       installlayout.Current,
		ScheduleMaintenance: maintenance.Schedule,
		Now:                 time.Now,
		HTTPClient:          &http.Client{Timeout: 30 * time.Second},
		RemoveExecutable:    selfmanage.RemoveExecutable,
		RollbackExecutable:  selfmanage.Rollback,
		UpgradeRelease:      upgrade.Execute,
		ReconcileInstalled:  reconcileInstalledResources,
	}
}

func isProbeInvocation(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		return true
	}
	return args[0] == "__self-test"
}

func Execute(info BuildInfo) error {
	if !isProbeInvocation(os.Args[1:]) {
		cleanup, err := registerCurrentInstance(info)
		if err != nil {
			return MutationError("register Buda payload instance", err)
		}
		defer cleanup()
	}
	deps := DefaultDependencies()
	deps.Version = info.Version
	root := NewRootCommand(deps, info, NewApplicationCommands(deps)...)
	err := root.Execute()
	if errors.Is(err, errHelpRendered) {
		return nil
	}
	if err != nil && deps.Options.JSON && !IsErrorRendered(err) {
		if renderErr := writeErrorJSON(deps.Err, err); renderErr == nil {
			return &renderedError{cause: err}
		}
	}
	return err
}

// IsErrorRendered reports whether Execute already emitted the one-document
// JSON diagnostic selected by --json.
func IsErrorRendered(err error) bool {
	var rendered *renderedError
	return errors.As(err, &rendered)
}

func writeErrorJSON(writer io.Writer, err error) error {
	body := errorBody{Code: ExitCode(err), Message: err.Error()}
	var qmdError *qmd.CommandError
	if errors.As(err, &qmdError) {
		copy := *qmdError
		body.QMD = &copy
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(errorDocument{Error: body})
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
	var helpTreeGlobalFlags bool
	var helpDepth string

	root := &cobra.Command{
		Use:           "buda",
		Short:         "Maintain and retrieve one explicit evidence-backed OKF wiki.",
		Example:       "  buda status --wiki C:\\knowledge\\team-wiki",
		Version:       info.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          NoArgs,
		PersistentPreRunE: func(command *cobra.Command, _ []string) error {
			if command.Flags().Changed("help-tree-depth") || helpTree || helpTreeGlobalFlags {
				if _, err := clihelp.ParseDepth(helpDepth); err != nil {
					return UsageError("%v", err)
				}
			}
			if helpDocs {
				markdown, err := clihelp.Markdown(command)
				if err != nil {
					return err
				}
				fmt.Fprint(command.OutOrStdout(), markdown)
				return errHelpRendered
			}
			if helpTree || command.Flags().Changed("help-tree-depth") {
				tree, err := clihelp.Tree(command, helpDepth, helpTreeGlobalFlags)
				if err != nil {
					return UsageError("%v", err)
				}
				fmt.Fprint(command.OutOrStdout(), tree)
				return errHelpRendered
			}
			if command == command.Root() && strings.TrimSpace(options.Wiki) != "" {
				repo, err := repository.Open(options.Wiki)
				if err != nil {
					return RepositoryError("open explicit --wiki", err)
				}
				options.Wiki = repo.Root
			} else if isRepositoryCommand(command) {
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
			if !maintenance.ShouldSchedule(command) {
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
			message := fmt.Sprintf("Hello Windows - buda %s", info.Version)
			selected := strings.TrimSpace(options.Wiki) != ""
			if options.JSON {
				output := map[string]any{"command": "buda", "version": info.Version, "wiki_selected": selected, "message": message}
				if selected {
					output["wiki"] = options.Wiki
				}
				return WriteJSON(command, output)
			}
			fmt.Fprintln(command.OutOrStdout(), message)
			if selected {
				fmt.Fprintf(command.OutOrStdout(), "wiki: %s\n", options.Wiki)
			}
			return nil
		},
	}
	root.SetIn(deps.In)
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)
	root.SetVersionTemplate("{{.Version}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpCommand(&cobra.Command{Use: "help", Hidden: true})
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return UsageError("parse flags: %v", err)
	})

	flags := root.PersistentFlags()
	flags.StringVar(&options.Wiki, "wiki", "", "Explicit path to the one wiki repository")
	flags.BoolVar(&options.JSON, "json", false, "Emit one stable JSON document")
	flags.BoolVar(&helpTree, "help-tree", false, "Show the public command subtree")
	flags.StringVar(&helpDepth, "help-tree-depth", "max", "Limit command-tree recursion to max or an integer greater than 1")
	flags.BoolVar(&helpTreeGlobalFlags, "help-tree-global-flags", false, "Repeat inherited global flags under every command in the help tree")
	flags.BoolVar(&helpDocs, "help-docs", false, "Show deterministic Markdown for this command scope")

	root.AddCommand(commands...)
	root.AddCommand(newAgentCommand(deps))
	root.AddCommand(newUpgradeCommand(deps, info))
	root.AddCommand(newUninstallCommand(deps))
	root.AddCommand(newMaintenanceCommand(deps))
	root.AddCommand(newSelfTestCommand(info))
	wrapDeveloperHelpArgs(root)
	return root
}

// newSelfTestCommand is intentionally hidden from the public command tree.
// Installers and lifecycle transactions use it to prove the candidate can
// construct its command surface and read embedded resources before activation.
func newSelfTestCommand(info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:    "__self-test",
		Hidden: true,
		Args:   NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if info.Version != "dev" && !releasecatalog.IsSemver(info.Version) {
				return fmt.Errorf("invalid build version %q", info.Version)
			}
			if _, err := fs.ReadFile(skills.FS, "guiho-s-0002-buda/SKILL.md"); err != nil {
				return fmt.Errorf("read embedded skill: %w", err)
			}
			if _, err := fs.ReadFile(prompts.FS, "guiho-i-buda.md"); err != nil {
				return fmt.Errorf("read embedded instruction: %w", err)
			}
			if _, err := fs.ReadFile(prompts.FS, "guiho-p-buda.md"); err != nil {
				return fmt.Errorf("read embedded prompt: %w", err)
			}
			if _, err := fs.ReadFile(schemas.FS, "buda.schema.json"); err != nil {
				return fmt.Errorf("read embedded project schema: %w", err)
			}
			if _, err := fs.ReadFile(schemas.FS, "buda.global.schema.json"); err != nil {
				return fmt.Errorf("read embedded global schema: %w", err)
			}
			if _, err := fs.ReadFile(examples.FS, "buda.example.yaml"); err != nil {
				return fmt.Errorf("read embedded project example: %w", err)
			}
			if _, err := fs.ReadFile(examples.FS, "buda.global.example.yaml"); err != nil {
				return fmt.Errorf("read embedded global example: %w", err)
			}
			fmt.Fprintln(command.OutOrStdout(), "ok")
			return nil
		},
	}
}

// wrapDeveloperHelpArgs lets the three generated developer-help forms inspect
// any command scope without satisfying that command's ordinary positionals.
// Cobra evaluates Args before PersistentPreRunE, so this must wrap the live
// tree rather than relying on the pre-run renderer alone.
func wrapDeveloperHelpArgs(command *cobra.Command) {
	original := command.Args
	command.Args = func(current *cobra.Command, args []string) error {
		if developerHelpRequested(current) {
			return nil
		}
		if original == nil {
			return nil
		}
		return original(current, args)
	}
	for _, child := range command.Commands() {
		wrapDeveloperHelpArgs(child)
	}
}

func developerHelpRequested(command *cobra.Command) bool {
	return command.Flags().Changed("help-tree") ||
		command.Flags().Changed("help-tree-depth") ||
		command.Flags().Changed("help-tree-global-flags") ||
		command.Flags().Changed("help-docs")
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
	if deps.Interactive == nil {
		reader := deps.In
		deps.Interactive = func() bool { return interactiveReader(reader) }
	}
	if deps.Executable == nil {
		deps.Executable = os.Executable
	}
	if deps.HomeDir == nil {
		deps.HomeDir = os.UserHomeDir
	}
	if deps.Agents == nil {
		deps.Agents = agent.NewService(agent.DefaultResources(), agent.WithHomeDir(deps.HomeDir))
	}
	if deps.InstallLayout == nil {
		deps.InstallLayout = installlayout.Current
	}
	if deps.ScheduleMaintenance == nil {
		deps.ScheduleMaintenance = maintenance.Schedule
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.HTTPClient == nil {
		deps.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if deps.RemoveExecutable == nil {
		deps.RemoveExecutable = selfmanage.RemoveExecutable
	}
	if deps.RollbackExecutable == nil {
		deps.RollbackExecutable = selfmanage.Rollback
	}
	if deps.UpgradeRelease == nil {
		deps.UpgradeRelease = upgrade.Execute
	}
	if deps.ReconcileInstalled == nil {
		deps.ReconcileInstalled = reconcileInstalledResources
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
	return top.Name() != "agent" && top.Name() != "upgrade" && top.Name() != "uninstall" && !top.Hidden
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

func dependencyNow(deps Dependencies) time.Time {
	if deps.Now != nil {
		return deps.Now().UTC()
	}
	return time.Now().UTC()
}

func releaseVersion(deps Dependencies) string {
	if value := strings.TrimSpace(deps.Version); value != "" {
		value = strings.TrimPrefix(value, "v")
		if releasecatalog.IsSemver(value) {
			return value
		}
	}
	return "0.2.0"
}
