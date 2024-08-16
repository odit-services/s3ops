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
