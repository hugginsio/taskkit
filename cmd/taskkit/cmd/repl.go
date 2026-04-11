// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive TaskKit shell",
	Long: `Start an interactive TaskKit shell.

All commands, flags, aliases, and autocomplete work exactly as they do from
the regular CLI. Type 'exit' or press Ctrl+D to quit.`,
	GroupID: CommandGroupUtility,
	// Override root's PersistentPreRunE/PostRun so the repl invocation itself
	// does not open the database. Each inner command handles its own lifecycle.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	RunE:              replRunE,
}

func init() {
	rootCmd.AddCommand(replCmd)
}

func replRunE(cmd *cobra.Command, _ []string) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "taskkit> ",
		HistoryFile:     replHistoryPath(),
		AutoComplete:    cobraCompleter{},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		return err
	}

	defer rl.Close()
	ctx := cmd.Context()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}

			continue
		}

		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			break
		}

		parts := strings.Fields(line)
		if parts[0] == "repl" {
			cmd.PrintErrln("error: recursive repl is not allowed")
			continue
		}

		resetFlags(rootCmd)
		rootCmd.SetArgs(parts)
		if err := rootCmd.ExecuteContext(ctx); err != nil {
			handleError(err)
		}
	}

	return nil
}

// resetFlags recursively resets all flags in cmd's tree to their default values.
// Cobra does not reset flag state between Execute calls, so this must be called
// before each dispatch in the REPL loop.
func resetFlags(cmd *cobra.Command) {
	for _, fs := range []*pflag.FlagSet{cmd.Flags(), cmd.PersistentFlags()} {
		fs.VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				f.Value.Set(f.DefValue) //nolint:errcheck
				f.Changed = false
			}
		})
	}

	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

// replHistoryPath returns the path to the readline history file.
func replHistoryPath() string {
	dir := filepath.Join(xdg.CacheHome, "taskkit")
	os.MkdirAll(dir, 0o700) //nolint:errcheck // readline degrades gracefully
	return filepath.Join(dir, "history")
}

// cobraCompleter implements readline.AutoCompleter using Cobra's __complete
// machinery, giving full parity with shell tab completion including any
// RegisterFlagCompletionFunc callbacks.
type cobraCompleter struct{}

func (cobraCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	text := string(line[:pos])

	// Determine the word currently being completed.
	var prefix string
	if len(text) > 0 && !strings.HasSuffix(text, " ") {
		parts := strings.Fields(text)
		if len(parts) > 0 {
			prefix = parts[len(parts)-1]
		}
	}

	// Build the completion query: everything before the current word + the prefix.
	beforePrefix := strings.TrimRight(text, " ")
	if prefix != "" {
		beforePrefix = text[:strings.LastIndex(text, prefix)]
	}

	prevParts := strings.Fields(beforePrefix)

	// Invoke cobra's __complete with the previous args and the prefix.
	completeArgs := append([]string{"__complete"}, prevParts...)
	completeArgs = append(completeArgs, prefix)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(io.Discard)
	resetFlags(rootCmd)
	rootCmd.SetArgs(completeArgs)
	rootCmd.Execute() //nolint:errcheck

	// Restore normal output.
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	// Parse completion output. Each line is "value\tDescription" or ":directive".
	var candidates [][]rune
	for l := range strings.SplitSeq(buf.String(), "\n") {
		if l == "" || strings.HasPrefix(l, ":") {
			continue
		}

		value := strings.SplitN(l, "\t", 2)[0]
		if strings.HasPrefix(value, prefix) {
			candidates = append(candidates, []rune(value[len(prefix):]))
		}
	}

	return candidates, len([]rune(prefix))
}
