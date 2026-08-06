package cli

import (
	"fmt"
	"strings"
)

// cmdCompletion prints a shell completion script for bash, zsh, or
// fish. Deliberately scoped to top-level subcommand names only — not
// per-command flags. Flag-aware completion means teaching the shell
// script about every FlagSet in internal/cli and keeping the two in
// sync forever; subcommand completion solves the actual discovery
// problem ("what commands does this have") with a static list that
// can't drift, since it's generated from the same commandNames slice
// Run() dispatches against.
func cmdCompletion(args []string) {
	if len(args) == 0 {
		usageError("completion requires a shell: bash|zsh|fish")
	}
	shell := args[0]

	var script string
	switch shell {
	case "bash":
		script = bashCompletion()
	case "zsh":
		script = zshCompletion()
	case "fish":
		script = fishCompletion()
	default:
		usageError("unknown shell: " + shell + " (want bash|zsh|fish)")
	}
	fmt.Print(script)
}

func bashCompletion() string {
	return `# devia bash completion
# Install:  source <(devia completion bash)
# Persist:  devia completion bash > /etc/bash_completion.d/devia
_devia_completions() {
  local cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "` + strings.Join(commandNames, " ") + `" -- "$cur") )
  fi
}
complete -F _devia_completions devia
`
}

func zshCompletion() string {
	return `#compdef devia
# devia zsh completion
# Install:  source <(devia completion zsh)
# Persist:  devia completion zsh > "${fpath[1]}/_devia"
_devia() {
  local -a commands
  commands=(` + strings.Join(commandNames, " ") + `)
  _describe 'command' commands
}
_devia
`
}

func fishCompletion() string {
	return `# devia fish completion
# Install:  devia completion fish | source
# Persist:  devia completion fish > ~/.config/fish/completions/devia.fish
complete -c devia -f -n "__fish_use_subcommand" -a "` + strings.Join(commandNames, " ") + `"
`
}
