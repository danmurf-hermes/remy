package persona_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danmurf/remy/internal/persona"
)

//nolint:gocyclo
func TestLoadPersona(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		check    func(t *testing.T, p *persona.Persona)
		wantErr  bool
		errType  any
	}{
		{
			name:     "valid persona with all frontmatter",
			filename: "test-persona.md",
			content: `---
provider: ollama
model: llama3.1:8b
temperature: 0.7
max_tokens: 4096
---
# Remy

You are a helpful assistant.
`,
			check: func(t *testing.T, p *persona.Persona) {
				if p.Name != "test-persona" {
					t.Errorf("name = %q, want %q", p.Name, "test-persona")
				}
				if p.Provider != "ollama" {
					t.Errorf("provider = %q, want %q", p.Provider, "ollama")
				}
				if p.Model != "llama3.1:8b" {
					t.Errorf("model = %q, want %q", p.Model, "llama3.1:8b")
				}
				if p.Temperature == nil || *p.Temperature != 0.7 {
					t.Errorf("temperature = %v, want 0.7", p.Temperature)
				}
				if p.MaxTokens == nil || *p.MaxTokens != 4096 {
					t.Errorf("max_tokens = %v, want 4096", p.MaxTokens)
				}
				if p.Body == "" {
					t.Error("body should not be empty")
				}
			},
		},
		{
			name:     "minimal persona (no frontmatter)",
			filename: "minimal.md",
			content: `# Remy

You are a helpful assistant.
`,
			check: func(t *testing.T, p *persona.Persona) {
				if p.Name != "minimal" {
					t.Errorf("name = %q, want %q", p.Name, "minimal")
				}
				if p.Provider != "" {
					t.Errorf("expected empty provider, got %q", p.Provider)
				}
				if p.Model != "" {
					t.Errorf("expected empty model, got %q", p.Model)
				}
				if p.Temperature != nil {
					t.Errorf("expected nil temperature, got %v", *p.Temperature)
				}
				if p.MaxTokens != nil {
					t.Errorf("expected nil max_tokens, got %v", *p.MaxTokens)
				}
				if p.Body == "" {
					t.Error("body should not be empty")
				}
			},
		},
		{
			name: "partial frontmatter",
			content: `---
provider: openai
---
# Remy

You are helpful.
`,
			check: func(t *testing.T, p *persona.Persona) {
				if p.Provider != "openai" {
					t.Errorf("provider = %q, want %q", p.Provider, "openai")
				}
				if p.Model != "" {
					t.Errorf("expected empty model, got %q", p.Model)
				}
				if p.Temperature != nil {
					t.Errorf("expected nil temperature")
				}
			},
		},
		{
			name:    "file not found",
			content: "",
			wantErr: true,
			errType: &persona.ErrPersonaNotFound{},
		},
		{
			name: "invalid temperature",
			content: `---
temperature: not-a-number
---
body
`,
			wantErr: true,
			errType: &persona.ErrInvalidFrontmatter{},
		},
		{
			name: "invalid max_tokens",
			content: `---
max_tokens: not-a-number
---
body
`,
			wantErr: true,
			errType: &persona.ErrInvalidFrontmatter{},
		},
		{
			name: "unclosed frontmatter",
			content: `---
provider: ollama
body
`,
			wantErr: true,
			errType: &persona.ErrInvalidFrontmatter{},
		},
		{
			name: "malformed frontmatter line",
			content: `---
badline
---
body
`,
			wantErr: true,
			errType: &persona.ErrInvalidFrontmatter{},
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
			errType: &persona.ErrPersonaNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var path string
			if tt.content != "" {
				// Use the specified filename or derive from test name
				filename := tt.filename
				if filename == "" {
					filename = sanitizeName(tt.name) + ".md"
				}
				path = filepath.Join(dir, filename)
				if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				path = filepath.Join(dir, "nonexistent.md")
			}

			p, err := persona.LoadPersona(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errType != nil {
					switch tt.errType.(type) {
					case *persona.ErrPersonaNotFound:
						var e *persona.ErrPersonaNotFound
						if !as(err, &e) {
							t.Fatalf("expected ErrPersonaNotFound, got %T", err)
						}
					case *persona.ErrInvalidFrontmatter:
						var e *persona.ErrInvalidFrontmatter
						if !as(err, &e) {
							t.Fatalf("expected ErrInvalidFrontmatter, got %T", err)
						}
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, p)
			}
		})
	}
}

func TestListPersonas(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		activeName string
		want       int
		wantActive string
	}{
		{
			name: "lists all persona files",
			files: map[string]string{
				"default.md": `---
---
# Default`,
				"creative.md": `---
---
# Creative`,
				"quick.md": `---
---
# Quick`,
			},
			activeName: "default",
			want:       3,
			wantActive: "default",
		},
		{
			name: "skips non-md files",
			files: map[string]string{
				"default.md": `---
---
# Default`,
				"notes.txt": "not a persona",
				"readme.md": `---
---
# Readme`,
			},
			activeName: "default",
			want:       2,
		},
		{
			name:  "empty directory",
			files: map[string]string{},
			want:  0,
		},
		{
			name:  "nonexistent directory",
			files: nil,
			want:  0,
		},
		{
			name: "skips invalid persona files",
			files: map[string]string{
				"valid.md": `---
---
# Valid`,
				"invalid.md": `---
bad
---
body`,
			},
			activeName: "valid",
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.files != nil {
				for name, content := range tt.files {
					path := filepath.Join(dir, name)
					if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				dir = filepath.Join(t.TempDir(), "nonexistent")
			}

			summaries, err := persona.ListPersonas(dir, tt.activeName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(summaries) != tt.want {
				t.Errorf("got %d summaries, want %d", len(summaries), tt.want)
			}
			if tt.wantActive != "" {
				found := false
				for _, s := range summaries {
					if s.Active {
						if s.Name != tt.wantActive {
							t.Errorf("active persona = %q, want %q", s.Name, tt.wantActive)
						}
						found = true
						break
					}
				}
				if !found {
					t.Error("no active persona found in summaries")
				}
			}
		})
	}
}

//nolint:gocyclo
func TestSavePersona(t *testing.T) {
	tests := []struct {
		name    string
		persona *persona.Persona
		check   func(t *testing.T, path string)
		wantErr bool
	}{
		{
			name: "saves persona with all fields",
			persona: &persona.Persona{
				Name:        "test",
				Provider:    "ollama",
				Model:       "llama3.1:8b",
				Temperature: float64Ptr(0.7),
				MaxTokens:   intPtr(4096),
				Body:        "# Test\n\nThis is a test persona.",
			},
			check: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				content := string(data)
				if !contains(content, "provider: ollama") {
					t.Error("missing provider in saved file")
				}
				if !contains(content, "model: llama3.1:8b") {
					t.Error("missing model in saved file")
				}
				if !contains(content, "temperature: 0.7") {
					t.Error("missing temperature in saved file")
				}
				if !contains(content, "max_tokens: 4096") {
					t.Error("missing max_tokens in saved file")
				}
				if !contains(content, "This is a test persona.") {
					t.Error("missing body in saved file")
				}
			},
		},
		{
			name: "saves minimal persona",
			persona: &persona.Persona{
				Name: "minimal",
				Body: "# Minimal\n\nJust a body.",
			},
			check: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				content := string(data)
				if contains(content, "provider:") {
					t.Error("should not contain provider")
				}
				if !contains(content, "Just a body.") {
					t.Error("missing body")
				}
			},
		},
		{
			name: "creates directory if needed",
			persona: &persona.Persona{
				Name: "deep",
				Body: "# Deep\n\nNested.",
			},
			check: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Fatal("file was not created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "subdir", tt.persona.Name+".md")

			err := persona.SavePersona(path, tt.persona)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, path)
			}
		})
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.md")

	original := &persona.Persona{
		Name:        "roundtrip",
		Provider:    "ollama",
		Model:       "llama3.1:8b",
		Temperature: float64Ptr(0.5),
		MaxTokens:   intPtr(2048),
		Body:        "# Roundtrip\n\nTesting save and load.",
	}

	if err := persona.SavePersona(path, original); err != nil {
		t.Fatalf("SavePersona: %v", err)
	}

	loaded, err := persona.LoadPersona(path)
	if err != nil {
		t.Fatalf("LoadPersona: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Provider != original.Provider {
		t.Errorf("provider = %q, want %q", loaded.Provider, original.Provider)
	}
	if loaded.Model != original.Model {
		t.Errorf("model = %q, want %q", loaded.Model, original.Model)
	}
	if loaded.Temperature == nil || *loaded.Temperature != *original.Temperature {
		t.Errorf("temperature = %v, want %v", loaded.Temperature, original.Temperature)
	}
	if loaded.MaxTokens == nil || *loaded.MaxTokens != *original.MaxTokens {
		t.Errorf("max_tokens = %v, want %v", loaded.MaxTokens, original.MaxTokens)
	}
	if loaded.Body != original.Body {
		t.Errorf("body = %q, want %q", loaded.Body, original.Body)
	}
}

func TestDir(t *testing.T) {
	dir, err := persona.Dir()
	if err != nil {
		t.Fatalf("Dir failed: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".remy", "personas")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

// Helpers

func sanitizeName(name string) string {
	// Replace spaces and special characters for use as a filename
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c == ' ' || c == '(' || c == ')' {
			result = append(result, '-')
		}
	}
	return string(result)
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func as(err error, target any) bool {
	switch t := target.(type) {
	case **persona.ErrPersonaNotFound:
		e, ok := err.(*persona.ErrPersonaNotFound)
		if ok {
			*t = e
			return true
		}
		return false
	case **persona.ErrInvalidFrontmatter:
		e, ok := err.(*persona.ErrInvalidFrontmatter)
		if ok {
			*t = e
			return true
		}
		return false
	}
	return false
}
