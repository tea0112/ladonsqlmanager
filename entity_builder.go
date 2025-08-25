package ladonsqlmanager

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/ladonsqlmanager/models"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
)

// EntityBuilder provides a fluent interface for building BaseEntity instances
type EntityBuilder struct {
	template   string
	startDelim byte
	endDelim   byte
	id         string
	compiled   string
	hasRegex   bool
	err        error
}

// NewEntityBuilder creates a new EntityBuilder instance
func NewEntityBuilder() *EntityBuilder {
	return &EntityBuilder{}
}

// WithTemplate sets the template and automatically sanitizes it
func (b *EntityBuilder) WithTemplate(template string) *EntityBuilder {
	if b.err != nil {
		return b
	}

	b.template = sanitizeTemplate(template)
	if b.template == "" {
		b.err = errors.New("template cannot be empty after sanitization")
		return b
	}

	return b
}

// WithDelimiters sets the start and end delimiters for regex compilation
func (b *EntityBuilder) WithDelimiters(startDelim, endDelim byte) *EntityBuilder {
	if b.err != nil {
		return b
	}

	b.startDelim = startDelim
	b.endDelim = endDelim
	return b
}

// GenerateID generates a SHA256-based ID from the template
func (b *EntityBuilder) GenerateID() *EntityBuilder {
	if b.err != nil {
		return b
	}

	if b.template == "" {
		b.err = errors.New("template must be set before generating ID")
		return b
	}

	h := sha256.New()
	_, _ = h.Write([]byte(b.template))
	b.id = fmt.Sprintf("%x", h.Sum(nil))

	return b
}

// CompileTemplate compiles the template using the provided delimiters
func (b *EntityBuilder) CompileTemplate() *EntityBuilder {
	if b.err != nil {
		return b
	}

	if b.template == "" {
		b.err = errors.New("template must be set before compilation")
		return b
	}

	compiled, err := compiler.CompileRegex(b.template, b.startDelim, b.endDelim)
	if err != nil {
		b.err = errors.WithStack(err)
		return b
	}

	b.compiled = compiled.String()
	b.hasRegex = strings.Index(b.template, string(b.startDelim)) >= 0

	return b
}

// WithCustomID allows setting a custom ID instead of generating one
func (b *EntityBuilder) WithCustomID(id string) *EntityBuilder {
	if b.err != nil {
		return b
	}

	if id == "" {
		b.err = errors.New("custom ID cannot be empty")
		return b
	}

	b.id = id
	return b
}

// Build creates the final BaseEntity instance
func (b *EntityBuilder) Build() (models.BaseEntity, error) {
	if b.err != nil {
		return models.BaseEntity{}, b.err
	}

	// Validate required fields
	if b.template == "" {
		return models.BaseEntity{}, errors.New("template is required")
	}
	if b.id == "" {
		return models.BaseEntity{}, errors.New("ID is required (call GenerateID() or WithCustomID())")
	}
	if b.compiled == "" {
		return models.BaseEntity{}, errors.New("compiled template is required (call CompileTemplate())")
	}

	baseEntity := models.BaseEntity{
		ID:       b.id,
		Template: b.template,
		Compiled: b.compiled,
		HasRegex: b.hasRegex,
	}

	// Validate the built entity
	if err := baseEntity.Validate(); err != nil {
		return models.BaseEntity{}, errors.WithStack(err)
	}

	return baseEntity, nil
}

// Reset resets the builder to its initial state for reuse
func (b *EntityBuilder) Reset() *EntityBuilder {
	b.template = ""
	b.startDelim = 0
	b.endDelim = 0
	b.id = ""
	b.compiled = ""
	b.hasRegex = false
	b.err = nil
	return b
}

// GetError returns any error that occurred during building
func (b *EntityBuilder) GetError() error {
	return b.err
}

// EntityBuilderDirector provides high-level methods for common building patterns
type EntityBuilderDirector struct {
	builder *EntityBuilder
}

// NewEntityBuilderDirector creates a new director with a builder
func NewEntityBuilderDirector() *EntityBuilderDirector {
	return &EntityBuilderDirector{
		builder: NewEntityBuilder(),
	}
}

// BuildStandardEntity builds a standard entity with template, delimiters, and auto-generated ID
func (d *EntityBuilderDirector) BuildStandardEntity(template string, startDelim, endDelim byte) (models.BaseEntity, error) {
	return d.builder.Reset().
		WithTemplate(template).
		WithDelimiters(startDelim, endDelim).
		GenerateID().
		CompileTemplate().
		Build()
}

// BuildEntityWithCustomID builds an entity with a custom ID
func (d *EntityBuilderDirector) BuildEntityWithCustomID(template string, id string, startDelim, endDelim byte) (models.BaseEntity, error) {
	return d.builder.Reset().
		WithTemplate(template).
		WithCustomID(id).
		WithDelimiters(startDelim, endDelim).
		CompileTemplate().
		Build()
}
