package app

import (
	"github.com/danmurf/remy/internal/persona"
)

// PersonaLoader wraps the persona package functions into the
// agent.PersonaLoader interface.
type PersonaLoader struct{}

// NewPersonaLoader creates a new PersonaLoader.
func NewPersonaLoader() *PersonaLoader {
	return &PersonaLoader{}
}

// LoadPersona loads a persona from the given file path.
func (l *PersonaLoader) LoadPersona(path string) (*persona.Persona, error) {
	return persona.LoadPersona(path)
}

// ListPersonas lists all personas in the given directory.
func (l *PersonaLoader) ListPersonas(dir, activeName string) ([]persona.Summary, error) {
	return persona.ListPersonas(dir, activeName)
}
