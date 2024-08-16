package v1alpha1

// Implement the condition enums
const (
	ConditionReady       = "Ready"
	ConditionFailed      = "Failed"
	ConditionReconciling = "Reconciling"
)

// Implement the reason enums
const (
	ReasonNotFound               = "NotFound"
	ReasonOffline                = "Offline"
	ReasonFinalizerFailedToApply = "FinalizerFailedToApply"
	ReasonRequestFailed          = "RequestFailed"
	ReasonCreateFailed           = "CreateFailed"
)

const (
	StateReconciling = "RECONCILING"
	StateFailed      = "FAILED"
	StatePending     = "PENDING"
	StateSuccess     = "SUCCESS"
)

const (
	ActionUnknown = "UNKNOWN"
	ActionCreate  = "CREATE"
	ActionUpdate  = "UPDATE"
	ActionDelete  = "DELETE"
)

type ServerReference struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" yaml:"name"`

	// +kubebuilder:validation:Required
	Namespace string `json:"namespace" yaml:"namespace"`
}

type CrStatus struct {
	State              string `json:"state,omitempty" yaml:"state,omitempty"`
	LastAction         string `json:"lastAction,omitempty" yaml:"lastAction,omitempty"`
	LastMessage        string `json:"lastMessage,omitempty" yaml:"lastMessage,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}
