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

type ServerReference struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" yaml:"name"`

	// +kubebuilder:validation:Required
	Namespace string `json:"namespace" yaml:"namespace"`
}
