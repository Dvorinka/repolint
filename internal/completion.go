package internal

import "fmt"

// Completion prints a shell completion script for the given shell.
func Completion(shell string) (string, error) {
	switch shell {
	case "bash":
		return `_repolint() {
  local cur prev
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=($(compgen -W "check fix version completion" -- "$cur"))
    return 0
  fi
  case "$prev" in
    --root|--config) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
    completion) COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur")); return 0 ;;
  esac
  COMPREPLY=($(compgen -W "--remote --root --config --dry-run --json" -- "$cur"))
}
complete -F _repolint repolint
`, nil
	case "zsh":
		return `#compdef repolint
_repolint() {
  local -a cmds=(check fix version completion)
  if (( CURRENT == 2 )); then
    _describe 'command' cmds
    return
  fi
  _arguments \
    '--remote[include gh-metadata checks]' \
    '--root[repo root]:dir:_files -/' \
    '--config[config file]:file:_files' \
    '--dry-run[print only]' \
    '--json[structured output]'
}
_repolint "$@"
`, nil
	case "fish":
		return `complete -c repolint -n '__fish_use_subcommand' -a 'check fix version completion'
complete -c repolint -l remote -d 'include gh-metadata checks'
complete -c repolint -l root -r -F -d 'repo root'
complete -c repolint -l config -r -F -d 'config file'
complete -c repolint -l dry-run -d 'print only'
complete -c repolint -l json -d 'structured output'
`, nil
	}
	return "", fmt.Errorf("unknown shell %q — use bash, zsh, or fish", shell)
}
