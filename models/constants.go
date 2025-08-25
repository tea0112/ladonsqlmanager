package models

// Constants for policy effects
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// Table name constants
const (
	TableNamePolicy            = "ladon_policy"
	TableNameSubject           = "ladon_subject"
	TableNameAction            = "ladon_action"
	TableNameResource          = "ladon_resource"
	TableNamePolicySubjectRel  = "ladon_policy_subject_rel"
	TableNamePolicyActionRel   = "ladon_policy_action_rel"
	TableNamePolicyResourceRel = "ladon_policy_resource_rel"
)

// Field size constants
const (
	PolicyIDMaxLength = 255
	EntityIDMaxLength = 64
	CompiledMaxLength = 511
	TemplateMaxLength = 511
)
