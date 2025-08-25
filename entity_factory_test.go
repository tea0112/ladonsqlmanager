// Additional test file for entity builder tests
package ladonsqlmanager

import (
	"testing"

	"github.com/ladonsqlmanager/models"
)

func TestSubjectFactory(t *testing.T) {
	factory := &SubjectFactory{}

	// Test entity creation
	baseEntity := models.BaseEntity{
		ID:       "test-id",
		Template: "user:*",
		Compiled: "user:.*",
		HasRegex: true,
	}

	entity := factory.CreateEntity(baseEntity)
	subject, ok := entity.(*models.Subject)
	if !ok {
		t.Fatalf("Expected *models.Subject, got %T", entity)
	}

	if subject.BaseEntity.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", subject.BaseEntity.ID)
	}

	// Test relationship creation
	relation := factory.CreateRelation("policy-1", "subject-1")
	subjectRel, ok := relation.(*models.PolicySubjectRel)
	if !ok {
		t.Fatalf("Expected *models.PolicySubjectRel, got %T", relation)
	}

	if subjectRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", subjectRel.Policy)
	}
	if subjectRel.Subject != "subject-1" {
		t.Errorf("Expected Subject 'subject-1', got '%s'", subjectRel.Subject)
	}

	// Test entity type
	if factory.GetEntityType() != itemTypeSubject {
		t.Errorf("Expected entity type '%s', got '%s'", itemTypeSubject, factory.GetEntityType())
	}
}

func TestActionFactory(t *testing.T) {
	factory := &ActionFactory{}

	// Test entity creation
	baseEntity := models.BaseEntity{
		ID:       "test-id",
		Template: "read",
		Compiled: "read",
		HasRegex: false,
	}

	entity := factory.CreateEntity(baseEntity)
	action, ok := entity.(*models.Action)
	if !ok {
		t.Fatalf("Expected *models.Action, got %T", entity)
	}

	if action.BaseEntity.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", action.BaseEntity.ID)
	}

	// Test relationship creation
	relation := factory.CreateRelation("policy-1", "action-1")
	actionRel, ok := relation.(*models.PolicyActionRel)
	if !ok {
		t.Fatalf("Expected *models.PolicyActionRel, got %T", relation)
	}

	if actionRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", actionRel.Policy)
	}
	if actionRel.Action != "action-1" {
		t.Errorf("Expected Action 'action-1', got '%s'", actionRel.Action)
	}

	// Test entity type
	if factory.GetEntityType() != itemTypeAction {
		t.Errorf("Expected entity type '%s', got '%s'", itemTypeAction, factory.GetEntityType())
	}
}

func TestResourceFactory(t *testing.T) {
	factory := &ResourceFactory{}

	// Test entity creation
	baseEntity := models.BaseEntity{
		ID:       "test-id",
		Template: "document:*",
		Compiled: "document:.*",
		HasRegex: true,
	}

	entity := factory.CreateEntity(baseEntity)
	resource, ok := entity.(*models.Resource)
	if !ok {
		t.Fatalf("Expected *models.Resource, got %T", entity)
	}

	if resource.BaseEntity.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", resource.BaseEntity.ID)
	}

	// Test relationship creation
	relation := factory.CreateRelation("policy-1", "resource-1")
	resourceRel, ok := relation.(*models.PolicyResourceRel)
	if !ok {
		t.Fatalf("Expected *models.PolicyResourceRel, got %T", relation)
	}

	if resourceRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", resourceRel.Policy)
	}
	if resourceRel.Resource != "resource-1" {
		t.Errorf("Expected Resource 'resource-1', got '%s'", resourceRel.Resource)
	}

	// Test entity type
	if factory.GetEntityType() != itemTypeResource {
		t.Errorf("Expected entity type '%s', got '%s'", itemTypeResource, factory.GetEntityType())
	}
}

func TestEntityFactoryRegistry(t *testing.T) {
	registry := NewEntityFactoryRegistry()

	// Test that default factories are registered
	supportedTypes := registry.GetSupportedTypes()
	expectedTypes := []string{itemTypeSubject, itemTypeAction, itemTypeResource}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}

	// Test getting existing factories
	subjectFactory, exists := registry.GetFactory(itemTypeSubject)
	if !exists {
		t.Error("Expected subject factory to exist")
	}
	if _, ok := subjectFactory.(*SubjectFactory); !ok {
		t.Errorf("Expected *SubjectFactory, got %T", subjectFactory)
	}

	actionFactory, exists := registry.GetFactory(itemTypeAction)
	if !exists {
		t.Error("Expected action factory to exist")
	}
	if _, ok := actionFactory.(*ActionFactory); !ok {
		t.Errorf("Expected *ActionFactory, got %T", actionFactory)
	}

	resourceFactory, exists := registry.GetFactory(itemTypeResource)
	if !exists {
		t.Error("Expected resource factory to exist")
	}
	if _, ok := resourceFactory.(*ResourceFactory); !ok {
		t.Errorf("Expected *ResourceFactory, got %T", resourceFactory)
	}

	// Test getting non-existing factory
	_, exists = registry.GetFactory("nonexistent")
	if exists {
		t.Error("Expected nonexistent factory to not exist")
	}

	// Test registering custom factory
	customFactory := &SubjectFactory{} // Using SubjectFactory as a mock
	registry.RegisterFactory("custom", customFactory)

	retrievedFactory, exists := registry.GetFactory("custom")
	if !exists {
		t.Error("Expected custom factory to exist after registration")
	}
	if retrievedFactory != customFactory {
		t.Error("Expected retrieved factory to be the same instance as registered")
	}
}

func TestSQLManager_createPolicyRelation(t *testing.T) {
	// Create a mock SQLManager with factory registry
	manager := &SQLManager{
		factoryRegistry: NewEntityFactoryRegistry(),
	}

	// Test subject relation creation
	factory, exists := manager.factoryRegistry.GetFactory(itemTypeSubject)
	if !exists {
		t.Fatal("Subject factory should exist")
	}

	relation := factory.CreateRelation("policy-1", "subject-1")
	subjectRel, ok := relation.(*models.PolicySubjectRel)
	if !ok {
		t.Fatalf("Expected *models.PolicySubjectRel, got %T", relation)
	}

	if subjectRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", subjectRel.Policy)
	}
	if subjectRel.Subject != "subject-1" {
		t.Errorf("Expected Subject 'subject-1', got '%s'", subjectRel.Subject)
	}

	// Test action relation creation
	factory, exists = manager.factoryRegistry.GetFactory(itemTypeAction)
	if !exists {
		t.Fatal("Action factory should exist")
	}

	relation = factory.CreateRelation("policy-1", "action-1")
	actionRel, ok := relation.(*models.PolicyActionRel)
	if !ok {
		t.Fatalf("Expected *models.PolicyActionRel, got %T", relation)
	}

	if actionRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", actionRel.Policy)
	}
	if actionRel.Action != "action-1" {
		t.Errorf("Expected Action 'action-1', got '%s'", actionRel.Action)
	}

	// Test resource relation creation
	factory, exists = manager.factoryRegistry.GetFactory(itemTypeResource)
	if !exists {
		t.Fatal("Resource factory should exist")
	}

	relation = factory.CreateRelation("policy-1", "resource-1")
	resourceRel, ok := relation.(*models.PolicyResourceRel)
	if !ok {
		t.Fatalf("Expected *models.PolicyResourceRel, got %T", relation)
	}

	if resourceRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", resourceRel.Policy)
	}
	if resourceRel.Resource != "resource-1" {
		t.Errorf("Expected Resource 'resource-1', got '%s'", resourceRel.Resource)
	}
}
