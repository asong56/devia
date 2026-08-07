package cli

import "testing"

func TestBashCompletion_ContainsAllCommands(t *testing.T) {
	script := bashCompletion()
	for _, name := range commandNames {
		if !containsWord(script, name) {
			t.Errorf("bash completion script missing command %q", name)
		}
	}
}

func TestZshCompletion_ContainsAllCommands(t *testing.T) {
	script := zshCompletion()
	for _, name := range commandNames {
		if !containsWord(script, name) {
			t.Errorf("zsh completion script missing command %q", name)
		}
	}
}

func TestFishCompletion_ContainsAllCommands(t *testing.T) {
	script := fishCompletion()
	for _, name := range commandNames {
		if !containsWord(script, name) {
			t.Errorf("fish completion script missing command %q", name)
		}
	}
}

func TestCompletionScripts_NonEmptyAndDistinct(t *testing.T) {
	bash := bashCompletion()
	zsh := zshCompletion()
	fish := fishCompletion()
	for name, s := range map[string]string{"bash": bash, "zsh": zsh, "fish": fish} {
		if s == "" {
			t.Errorf("%s completion script is empty", name)
		}
	}
	if bash == zsh || bash == fish || zsh == fish {
		t.Error("the three shell completion scripts should not be byte-identical to each other")
	}
}

func TestCmdCompletion_ValidShellsDoNotExit(t *testing.T) {
	// cmdCompletion only calls os.Exit (via usageError) on an unknown
	// or missing shell argument — the three valid shells just fmt.Print
	// and return normally, so this is safe to call in-process.
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			cmdCompletion([]string{shell})
		})
	}
}

// containsWord is a small helper avoiding strings.Contains's substring
// pitfall for this specific check: command names in devia are short
// and some are prefixes of others in principle (none currently are,
// but "json" being naively substring-matched inside a longer future
// name would be exactly the kind of false pass this guards against).
// It checks for the name appearing as a standalone token, delimited by
// anything that isn't a lowercase letter — not just whitespace, since
// every generated script immediately follows the last command name
// with a closing quote character (`...version"`) rather than a space,
// which a whitespace-only delimiter set would miss.
func containsWord(s, word string) bool {
	isWordChar := func(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') }
	inWord := false
	start := 0
	for i := 0; i <= len(s); i++ {
		atEnd := i == len(s)
		wordChar := !atEnd && isWordChar(s[i])
		if wordChar {
			if !inWord {
				start = i
				inWord = true
			}
			continue
		}
		if inWord {
			if s[start:i] == word {
				return true
			}
			inWord = false
		}
	}
	return false
}
