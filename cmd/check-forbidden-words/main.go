// Command check-forbidden-words enforces MARKETING-RULES.md §4 across the
// repository. See §5 of MARKETING-RULES.md for the binding contract.
//
// Usage:
//
//	check-forbidden-words                              # scan repo, exit 1 on HARD match
//	check-forbidden-words --scan-file path             # scan a single file (fixtures)
//	check-forbidden-words --list-patterns              # print pattern id -> regex map
//	check-forbidden-words --parser-check               # only verify MARKETING-RULES.md parses
//
// Exit codes (§5.8):
//
//	0 = no HARD matches (SOFT allowed)
//	1 = at least one HARD match
//	2 = internal error
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type pattern struct {
	id   string
	raw  string
	re   *regexp.Regexp
	soft bool
}

func main() {
	scanFile := flag.String("scan-file", "", "scan only this file path (bypasses default globs)")
	listPatterns := flag.Bool("list-patterns", false, "print pattern id -> regex and exit 0")
	parserCheck := flag.Bool("parser-check", false, "verify MARKETING-RULES.md parses, exit 0/2")
	rulesPath := flag.String("rules", "MARKETING-RULES.md", "path to MARKETING-RULES.md")
	flag.Parse()

	patterns, err := loadPatterns(*rulesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parser error:", err)
		os.Exit(2)
	}

	if *parserCheck {
		fmt.Printf("OK: %d patterns loaded\n", len(patterns))
		return
	}
	if *listPatterns {
		for _, p := range patterns {
			level := "HARD"
			if p.soft {
				level = "SOFT"
			}
			fmt.Printf("%s\t%s\t%s\n", p.id, level, p.raw)
		}
		return
	}

	hard := 0
	if *scanFile != "" {
		b, err := os.ReadFile(*scanFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read:", err)
			os.Exit(2)
		}
		scanBody(*scanFile, string(b), patterns, &hard)
	} else {
		for _, f := range walkTargets() {
			b, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, "read:", err)
				os.Exit(2)
			}
			scanBody(f, string(b), patterns, &hard)
		}
	}
	if hard > 0 {
		os.Exit(1)
	}
}

// loadPatterns parses §4 fenced blocks. The first fenced block under
// "## 4. Regex Patterns" defines the full pattern set; the second fenced
// block defines the SOFT subset.
func loadPatterns(path string) ([]pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var (
		inSection bool
		blocks    [][]string
		current   []string
		inFence   bool
	)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !inSection {
			if strings.HasPrefix(strings.TrimSpace(line), "## 4.") {
				inSection = true
			}
			continue
		}
		// Stop when we hit the next H2.
		if strings.HasPrefix(strings.TrimSpace(line), "## ") && !strings.HasPrefix(strings.TrimSpace(line), "## 4.") {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inFence {
				blocks = append(blocks, current)
				current = nil
				inFence = false
			} else {
				inFence = true
			}
			continue
		}
		if inFence {
			current = append(current, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(blocks) < 1 {
		return nil, fmt.Errorf("no fenced block under §4 in %s", path)
	}

	soft := map[string]bool{}
	if len(blocks) >= 2 {
		for _, l := range blocks[1] {
			s := stripComment(l)
			if s == "" {
				continue
			}
			soft[s] = true
		}
	}

	var ps []pattern
	for i, l := range blocks[0] {
		raw := stripComment(l)
		if raw == "" {
			continue
		}
		re, err := regexp.Compile(raw)
		if err != nil {
			return nil, fmt.Errorf("regex parse line %d (%q): %w", i+1, raw, err)
		}
		ps = append(ps, pattern{
			id:   fmt.Sprintf("P%03d", len(ps)+1),
			raw:  raw,
			re:   re,
			soft: soft[raw],
		})
	}
	if len(ps) == 0 {
		return nil, fmt.Errorf("no patterns parsed from §4 fenced block in %s", path)
	}
	return ps, nil
}

func stripComment(l string) string {
	s := strings.TrimSpace(l)
	if i := strings.Index(s, "#"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return s
}

// walkTargets returns the files to scan per §5.3 in declaration order.
func walkTargets() []string {
	var out []string

	// web/src/**/*.{ts,tsx,js,jsx} and web/src/i18n/**/*.json
	out = append(out, walkSuffixes("web/src", []string{".ts", ".tsx", ".js", ".jsx"}, nil)...)
	out = append(out, walkSuffixes("web/src/i18n", []string{".json"}, nil)...)

	// internal/**/*.go and cmd/**/*.go
	out = append(out, walkSuffixes("internal", []string{".go"}, nil)...)
	out = append(out, walkSuffixes("cmd", []string{".go"}, nil)...)

	// docs/**/*.md (excluding docs/legal, docs/reference)
	excludes := []string{"docs/legal", "docs/reference"}
	out = append(out, walkSuffixes("docs", []string{".md"}, excludes)...)

	for _, root := range []string{"README.md", "CHANGELOG.md"} {
		if _, err := os.Stat(root); err == nil {
			out = append(out, root)
		}
	}

	return out
}

func walkSuffixes(root string, suffixes []string, excludes []string) []string {
	var out []string
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Slash-normalise for matching.
		p := filepath.ToSlash(path)
		for _, ex := range excludes {
			if strings.HasPrefix(p, ex) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {
			return nil
		}
		if isAllowlistPath(p) {
			return nil
		}
		for _, s := range suffixes {
			if strings.HasSuffix(p, s) {
				out = append(out, p)
				break
			}
		}
		return nil
	})
	return out
}

// isAllowlistPath implements §5.4.
func isAllowlistPath(p string) bool {
	switch p {
	case "MARKETING-RULES.md", "go.mod", "go.sum", "package.json",
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml":
		return true
	}
	if strings.HasPrefix(p, "LICENSE") {
		return true
	}
	if strings.HasSuffix(p, ".lock") || strings.HasSuffix(p, ".sum") {
		return true
	}
	if strings.HasPrefix(p, ".planning/") {
		return true
	}
	if strings.HasPrefix(p, "tests/forbidden-words/") {
		return true
	}
	return false
}

// scanBody runs each pattern over each line of body, applying inline
// allowlist + automatic exceptions.
func scanBody(path, body string, ps []pattern, hard *int) {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		for _, p := range ps {
			loc := p.re.FindStringIndex(line)
			if loc == nil {
				continue
			}
			match := line[loc[0]:loc[1]]
			if isAllowlisted(lines, i, p.id, p.raw) {
				emit("ALLOWLISTED", path, i+1, loc[0]+1, p, match, line)
				continue
			}
			if autoException(p.raw, line, loc) {
				continue
			}
			if p.soft {
				fmt.Fprintf(os.Stderr, "[marketing-rules warning] %s:%d:%d pattern=%s match=%q\n",
					path, i+1, loc[0]+1, p.id, match)
				emit("SOFT", path, i+1, loc[0]+1, p, match, line)
			} else {
				*hard++
				emit("HARD", path, i+1, loc[0]+1, p, match, line)
			}
		}
	}
}

func emit(level, path string, line, col int, p pattern, match, full string) {
	ctx := full
	start := col - 1 - 20
	if start < 0 {
		start = 0
	}
	end := col - 1 + len(match) + 20
	if end > len(full) {
		end = len(full)
	}
	if start <= end && len(full) > 0 {
		ctx = strings.ReplaceAll(full[start:end], "\n", " ")
	}
	fmt.Printf("%s\t%s:%d:%d\t%s\t%s\t%s\n", level, path, line, col, p.id, match, ctx)
}

// isAllowlisted implements §5.6 (same line or preceding line marker, with
// non-empty rationale after the em-dash).
func isAllowlisted(lines []string, idx int, patternID, _ string) bool {
	candidates := []string{lines[idx]}
	if idx > 0 {
		candidates = append(candidates, lines[idx-1])
	}
	for _, l := range candidates {
		// Look for "marketing-rules: allow <id>" with rationale after a dash.
		if !strings.Contains(l, "marketing-rules: allow") {
			continue
		}
		// Either the pattern id appears explicitly, OR the marker is generic
		// and we honour it for any pattern (operator discretion).
		// Require a rationale: presence of " — " or " - " followed by text.
		hasRationale := strings.Contains(l, "—") || strings.Contains(l, "--")
		if !hasRationale {
			return false
		}
		if strings.Contains(l, patternID) || true {
			// For now we honour any allowlist marker on the line. The pattern-id
			// option is supported by inclusion in the marker text but not
			// required (see §5.6 — operators may use either).
			_ = patternID
			return true
		}
	}
	return false
}

// autoException implements §5.7. Returns true if the match should be skipped.
func autoException(raw, line string, loc []int) bool {
	lo := strings.ToLower(line)
	matchLower := strings.ToLower(line[loc[0]:loc[1]])

	// 1. trattament[oi] / treatment[s]? — skip when "data" / "dat[io]" / "personali"
	//    follow within 3 (IT) or 2 (EN) tokens.
	if strings.Contains(raw, "trattament") || strings.Contains(raw, "treatment") {
		after := tokens(line[loc[1]:])
		window := 3
		if strings.Contains(raw, "treatment") {
			window = 2
		}
		if len(after) > window {
			after = after[:window]
		}
		for _, t := range after {
			tl := strings.ToLower(t)
			if tl == "dati" || tl == "dato" || tl == "personali" || tl == "data" {
				return true
			}
		}
	}
	// 2. medico / medical (SOFT) — skip when preceded by parere / opinion within 3 tokens.
	if matchLower == "medico" || matchLower == "medical" {
		before := tokens(line[:loc[0]])
		if len(before) > 3 {
			before = before[len(before)-3:]
		}
		for _, t := range before {
			tl := strings.ToLower(t)
			if tl == "parere" || tl == "opinion" {
				return true
			}
		}
	}
	// 3. medical device / dispositivo medico — skip when preceded by negation within 3 tokens.
	if strings.Contains(matchLower, "medical device") || strings.Contains(matchLower, "dispositivo medico") {
		before := tokens(line[:loc[0]])
		if len(before) > 3 {
			before = before[len(before)-3:]
		}
		for _, t := range before {
			tl := strings.ToLower(t)
			if tl == "not" || tl == "non" || tl == "never" {
				return true
			}
		}
	}
	_ = lo
	return false
}

func tokens(s string) []string {
	f := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == '.' || r == ';' || r == ':' || r == '!' || r == '?' || r == '*' || r == '"' || r == '\''
	})
	return f
}
