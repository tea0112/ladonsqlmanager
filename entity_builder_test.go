package ladonsqlmanager

import (
	"strings"
	"testing"
)

func TestEntityBuilder_BasicFlow(t *testing.T) {
	builder := NewEntityBuilder()

	entity, err := builder.
		WithTemplate("user:admin").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entity.Template != "user:admin" {
		t.Errorf("Expected template 'user:admin', got '%s'", entity.Template)
	}

	if entity.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if entity.Compiled == "" {
		t.Error("Expected compiled template to be set")
	}

	if entity.HasRegex {
		t.Error("Expected HasRegex to be false for template without delimiters")
	}
}

func TestEntityBuilder_WithRegex(t *testing.T) {
	builder := NewEntityBuilder()

	entity, err := builder.
		WithTemplate("user:<.*>").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !entity.HasRegex {
		t.Error("Expected HasRegex to be true for template with delimiters")
	}
}

func TestEntityBuilder_CustomID(t *testing.T) {
	builder := NewEntityBuilder()
	customID := "custom-id-123"

	entity, err := builder.
		WithTemplate("user:admin").
		WithCustomID(customID).
		WithDelimiters('<', '>').
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entity.ID != customID {
		t.Errorf("Expected ID '%s', got '%s'", customID, entity.ID)
	}
}

func TestEntityBuilder_ErrorHandling(t *testing.T) {
	builder := NewEntityBuilder()

	// Test empty template
	_, err := builder.
		WithTemplate("").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err == nil {
		t.Error("Expected error for empty template")
	}

	// Test missing ID
	builder.Reset()
	_, err = builder.
		WithTemplate("user:admin").
		WithDelimiters('<', '>').
		CompileTemplate().
		Build()

	if err == nil {
		t.Error("Expected error for missing ID")
	}

	// Test missing compilation
	builder.Reset()
	_, err = builder.
		WithTemplate("user:admin").
		WithDelimiters('<', '>').
		GenerateID().
		Build()

	if err == nil {
		t.Error("Expected error for missing compilation")
	}
}

func TestEntityBuilder_Reset(t *testing.T) {
	builder := NewEntityBuilder()

	// Build first entity
	_, err := builder.
		WithTemplate("user:admin").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Reset and build second entity
	entity2, err := builder.Reset().
		WithTemplate("user:guest").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error after reset, got %v", err)
	}

	if entity2.Template != "user:guest" {
		t.Errorf("Expected template 'user:guest' after reset, got '%s'", entity2.Template)
	}
}

func TestEntityBuilder_ErrorPropagation(t *testing.T) {
	builder := NewEntityBuilder()

	// Introduce an error early in the chain
	builder.WithTemplate("").WithDelimiters('<', '>')

	// Continue the chain - error should propagate
	entity, err := builder.GenerateID().CompileTemplate().Build()

	if err == nil {
		t.Error("Expected error to propagate through the chain")
	}

	// Check that the entity is empty when there's an error
	if entity.ID != "" || entity.Template != "" {
		t.Error("Expected empty entity when there's an error")
	}
}

func TestEntityBuilderDirector(t *testing.T) {
	director := NewEntityBuilderDirector()

	// Test standard entity building
	entity, err := director.BuildStandardEntity("user:admin", '<', '>')
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entity.Template != "user:admin" {
		t.Errorf("Expected template 'user:admin', got '%s'", entity.Template)
	}

	if entity.ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Test entity with custom ID
	customID := "custom-123"
	entity2, err := director.BuildEntityWithCustomID("user:guest", customID, '<', '>')
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entity2.ID != customID {
		t.Errorf("Expected ID '%s', got '%s'", customID, entity2.ID)
	}

	if entity2.Template != "user:guest" {
		t.Errorf("Expected template 'user:guest', got '%s'", entity2.Template)
	}
}

func TestEntityBuilder_TemplateSanitization(t *testing.T) {
	builder := NewEntityBuilder()

	// Test template with whitespace
	entity, err := builder.
		WithTemplate("  user:admin  ").
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entity.Template != "user:admin" {
		t.Errorf("Expected sanitized template 'user:admin', got '%s'", entity.Template)
	}
}

func TestEntityBuilder_Validation(t *testing.T) {
	builder := NewEntityBuilder()

	// Create an entity that would fail BaseEntity validation
	// (using a template that's too long)
	longTemplate := strings.Repeat("a", 600) // Exceeds TemplateMaxLength

	_, err := builder.
		WithTemplate(longTemplate).
		WithDelimiters('<', '>').
		GenerateID().
		CompileTemplate().
		Build()

	if err == nil {
		t.Error("Expected validation error for template that's too long")
	}
}
