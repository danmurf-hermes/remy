// Package persona provides loading, parsing, saving, and switching of
// agent personas. Each persona is a Markdown file with YAML frontmatter
// that defines the agent's identity, tone, behavior, and optional model
// overrides.
package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Persona represents a loaded persona with parsed frontmatter and body.
type Persona struct {
	Name        string
	Provider    string
	Model       string
	Temperature *float64
	MaxTokens   *int
	Body        string
}

// Summary is a lightweight representation of a persona for listing.
type Summary struct {
	Name     string
	Provider string
	Model    string
	Active   bool
}

// ErrPersonaNotFound is returned when a persona file does not exist.
type ErrPersonaNotFound struct {
	Name string
}

func (e *ErrPersonaNotFound) Error() string {
	return fmt.Sprintf("persona %q not found", e.Name)
}

// ErrInvalidFrontmatter is returned when a persona file has invalid YAML frontmatter.
type ErrInvalidFrontmatter struct {
	Path string
	Err  string
}

func (e *ErrInvalidFrontmatter) Error() string {
	return fmt.Sprintf("invalid frontmatter in %q: %s", e.Path, e.Err)
}

// Dir returns the default persona directory (~/.remy/personas/).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".remy", "personas"), nil
}

// LoadPersona reads and parses a persona Markdown file. The file must have
// YAML frontmatter delimited by `---` lines, followed by the persona body.
func LoadPersona(path string) (*Persona, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is user-provided persona path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrPersonaNotFound{Name: filepath.Base(path)}
		}
		return nil, fmt.Errorf("reading persona file %q: %w", path, err)
	}

	content := string(data)
	frontmatter, body, err := parseFrontmatter(content)
	if err != nil {
		return nil, &ErrInvalidFrontmatter{Path: path, Err: err.Error()}
	}

	name := strings.TrimSuffix(filepath.Base(path), ".md")
	p := &Persona{
		Name: name,
		Body: strings.TrimSpace(body),
	}

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			return nil, &ErrInvalidFrontmatter{Path: path, Err: fmt.Sprintf("malformed line: %q", line)}
		}
		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "provider":
			p.Provider = value
		case "model":
			p.Model = value
		case "temperature":
			t, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, &ErrInvalidFrontmatter{Path: path, Err: fmt.Sprintf("invalid temperature %q: %v", value, err)}
			}
			p.Temperature = &t
		case "max_tokens":
			m, err := strconv.Atoi(value)
			if err != nil {
				return nil, &ErrInvalidFrontmatter{Path: path, Err: fmt.Sprintf("invalid max_tokens %q: %v", value, err)}
			}
			p.MaxTokens = &m
		}
	}

	return p, nil
}

// ListPersonas scans the given directory for Markdown files with YAML
// frontmatter and returns a summary of each. The active persona name is
// used to mark the active one.
func ListPersonas(dir, activeName string) ([]Summary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading persona directory %q: %w", dir, err)
	}

	var summaries []Summary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		p, err := LoadPersona(path)
		if err != nil {
			continue
		}
		summaries = append(summaries, Summary{
			Name:     p.Name,
			Provider: p.Provider,
			Model:    p.Model,
			Active:   p.Name == activeName,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries, nil
}

// SavePersona writes a persona to a Markdown file at the given path.
func SavePersona(path string, p *Persona) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating persona directory: %w", err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	if p.Provider != "" {
		fmt.Fprintf(&b, "provider: %s\n", p.Provider)
	}
	if p.Model != "" {
		fmt.Fprintf(&b, "model: %s\n", p.Model)
	}
	if p.Temperature != nil {
		fmt.Fprintf(&b, "temperature: %s\n", formatFloat(*p.Temperature))
	}
	if p.MaxTokens != nil {
		fmt.Fprintf(&b, "max_tokens: %d\n", *p.MaxTokens)
	}
	b.WriteString("---\n")
	b.WriteString(p.Body)
	b.WriteString("\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("writing persona file %q: %w", path, err)
	}
	return nil
}

// parseFrontmatter extracts YAML frontmatter and body from a Markdown file.
// The frontmatter is delimited by `---` lines at the start of the file.
func parseFrontmatter(content string) (frontmatter, body string, err error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return "", content, nil
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return "", content, fmt.Errorf("unclosed frontmatter: no closing '---' found")
	}

	frontmatter = strings.TrimSpace(rest[:endIdx])
	body = strings.TrimSpace(rest[endIdx+4:])
	return frontmatter, body, nil
}

func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return s
}
