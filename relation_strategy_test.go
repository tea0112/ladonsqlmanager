package ladonsqlmanager

import (
	"testing"

	"github.com/ladonsqlmanager/models"
)

func TestSubjectRelationStrategy(t *testing.T) {
	strategy := &SubjectRelationStrategy{}

	// Test relation creation
	relation := strategy.CreateRelation("policy-1", "subject-1")
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

	// Test relation type
	if strategy.GetRelationType() != itemTypeSubject {
		t.Errorf("Expected relation type '%s', got '%s'", itemTypeSubject, strategy.GetRelationType())
	}
}

func TestActionRelationStrategy(t *testing.T) {
	strategy := &ActionRelationStrategy{}

	// Test relation creation
	relation := strategy.CreateRelation("policy-1", "action-1")
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

	// Test relation type
	if strategy.GetRelationType() != itemTypeAction {
		t.Errorf("Expected relation type '%s', got '%s'", itemTypeAction, strategy.GetRelationType())
	}
}

func TestResourceRelationStrategy(t *testing.T) {
	strategy := &ResourceRelationStrategy{}

	// Test relation creation
	relation := strategy.CreateRelation("policy-1", "resource-1")
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

	// Test relation type
	if strategy.GetRelationType() != itemTypeResource {
		t.Errorf("Expected relation type '%s', got '%s'", itemTypeResource, strategy.GetRelationType())
	}
}

func TestRelationStrategyRegistry(t *testing.T) {
	registry := NewRelationStrategyRegistry()

	// Test that default strategies are registered
	supportedTypes := registry.GetSupportedTypes()
	expectedTypes := []string{itemTypeSubject, itemTypeAction, itemTypeResource}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}

	// Test getting existing strategies
	subjectStrategy, exists := registry.GetStrategy(itemTypeSubject)
	if !exists {
		t.Error("Expected subject strategy to exist")
	}
	if _, ok := subjectStrategy.(*SubjectRelationStrategy); !ok {
		t.Errorf("Expected *SubjectRelationStrategy, got %T", subjectStrategy)
	}

	actionStrategy, exists := registry.GetStrategy(itemTypeAction)
	if !exists {
		t.Error("Expected action strategy to exist")
	}
	if _, ok := actionStrategy.(*ActionRelationStrategy); !ok {
		t.Errorf("Expected *ActionRelationStrategy, got %T", actionStrategy)
	}

	resourceStrategy, exists := registry.GetStrategy(itemTypeResource)
	if !exists {
		t.Error("Expected resource strategy to exist")
	}
	if _, ok := resourceStrategy.(*ResourceRelationStrategy); !ok {
		t.Errorf("Expected *ResourceRelationStrategy, got %T", resourceStrategy)
	}

	// Test getting non-existing strategy
	_, exists = registry.GetStrategy("nonexistent")
	if exists {
		t.Error("Expected nonexistent strategy to not exist")
	}

	// Test registering custom strategy
	customStrategy := &SubjectRelationStrategy{} // Using SubjectRelationStrategy as a mock
	registry.RegisterStrategy("custom", customStrategy)

	retrievedStrategy, exists := registry.GetStrategy("custom")
	if !exists {
		t.Error("Expected custom strategy to exist after registration")
	}
	if retrievedStrategy != customStrategy {
		t.Error("Expected retrieved strategy to be the same instance as registered")
	}
}

func TestRelationContext(t *testing.T) {
	strategy := &SubjectRelationStrategy{}
	context := NewRelationContext(strategy)

	// Test relation creation through context
	relation := context.CreateRelation("policy-1", "subject-1")
	subjectRel, ok := relation.(*models.PolicySubjectRel)
	if !ok {
		t.Fatalf("Expected *models.PolicySubjectRel, got %T", relation)
	}

	if subjectRel.Policy != "policy-1" {
		t.Errorf("Expected Policy 'policy-1', got '%s'", subjectRel.Policy)
	}

	// Test strategy change
	newStrategy := &ActionRelationStrategy{}
	context.SetStrategy(newStrategy)

	relation2 := context.CreateRelation("policy-2", "action-1")
	actionRel, ok := relation2.(*models.PolicyActionRel)
	if !ok {
		t.Fatalf("Expected *models.PolicyActionRel after strategy change, got %T", relation2)
	}

	if actionRel.Policy != "policy-2" {
		t.Errorf("Expected Policy 'policy-2', got '%s'", actionRel.Policy)
	}
}

func TestRelationTypeDetector(t *testing.T) {
	registry := NewRelationStrategyRegistry()
	detector := NewRelationTypeDetector(registry)

	// Test subject relation detection
	subjectRel := &models.PolicySubjectRel{Policy: "policy-1", Subject: "subject-1"}
	strategy, err := detector.DetectAndGetStrategy(subjectRel)
	if err != nil {
		t.Fatalf("Expected no error for subject relation, got %v", err)
	}
	if _, ok := strategy.(*SubjectRelationStrategy); !ok {
		t.Errorf("Expected *SubjectRelationStrategy, got %T", strategy)
	}

	// Test action relation detection
	actionRel := &models.PolicyActionRel{Policy: "policy-1", Action: "action-1"}
	strategy, err = detector.DetectAndGetStrategy(actionRel)
	if err != nil {
		t.Fatalf("Expected no error for action relation, got %v", err)
	}
	if _, ok := strategy.(*ActionRelationStrategy); !ok {
		t.Errorf("Expected *ActionRelationStrategy, got %T", strategy)
	}

	// Test resource relation detection
	resourceRel := &models.PolicyResourceRel{Policy: "policy-1", Resource: "resource-1"}
	strategy, err = detector.DetectAndGetStrategy(resourceRel)
	if err != nil {
		t.Fatalf("Expected no error for resource relation, got %v", err)
	}
	if _, ok := strategy.(*ResourceRelationStrategy); !ok {
		t.Errorf("Expected *ResourceRelationStrategy, got %T", strategy)
	}

	// Test invalid relation type
	invalidRel := "invalid-relation"
	_, err = detector.DetectAndGetStrategy(invalidRel)
	if err == nil {
		t.Error("Expected error for invalid relation type")
	}
	if err != ErrInvalidRelationType {
		t.Errorf("Expected ErrInvalidRelationType, got %v", err)
	}
}

func TestEnhancedSQLManager_createPolicyRelationOptimized(t *testing.T) {
	// Create a mock SQLManager with strategy components
	manager := &SQLManager{
		strategyRegistry: NewRelationStrategyRegistry(),
	}
	manager.typeDetector = NewRelationTypeDetector(manager.strategyRegistry)

	// Test subject relation
	subjectRel := &models.PolicySubjectRel{Policy: "policy-1", Subject: "subject-1"}
	strategy, err := manager.typeDetector.DetectAndGetStrategy(subjectRel)
	if err != nil {
		t.Fatalf("Expected no error for subject relation detection, got %v", err)
	}
	if _, ok := strategy.(*SubjectRelationStrategy); !ok {
		t.Errorf("Expected *SubjectRelationStrategy, got %T", strategy)
	}

	// Test action relation
	actionRel := &models.PolicyActionRel{Policy: "policy-1", Action: "action-1"}
	strategy, err = manager.typeDetector.DetectAndGetStrategy(actionRel)
	if err != nil {
		t.Fatalf("Expected no error for action relation detection, got %v", err)
	}
	if _, ok := strategy.(*ActionRelationStrategy); !ok {
		t.Errorf("Expected *ActionRelationStrategy, got %T", strategy)
	}

	// Test resource relation
	resourceRel := &models.PolicyResourceRel{Policy: "policy-1", Resource: "resource-1"}
	strategy, err = manager.typeDetector.DetectAndGetStrategy(resourceRel)
	if err != nil {
		t.Fatalf("Expected no error for resource relation detection, got %v", err)
	}
	if _, ok := strategy.(*ResourceRelationStrategy); !ok {
		t.Errorf("Expected *ResourceRelationStrategy, got %T", strategy)
	}

	// Test invalid relation
	_, err = manager.typeDetector.DetectAndGetStrategy("invalid")
	if err == nil {
		t.Error("Expected error for invalid relation type")
	}
}
