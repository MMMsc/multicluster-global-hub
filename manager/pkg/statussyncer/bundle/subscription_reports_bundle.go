package bundle

import (
	appsv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/v1alpha1"

	"github.com/stolostron/multicluster-global-hub/pkg/bundle/status"
)

// NewSubscriptionReportsBundle creates a new instance of SubscriptionReportsBundle.
func NewSubscriptionReportsBundle() status.Bundle {
	return &SubscriptionReportsBundle{}
}

// SubscriptionReportsBundle abstracts management of subscription-reports bundle.
type SubscriptionReportsBundle struct {
	baseBundle
	Objects []*appsv1alpha1.SubscriptionReport `json:"objects"`
}

// GetObjects return all the objects that the bundle holds.
func (bundle *SubscriptionReportsBundle) GetObjects() []interface{} {
	result := make([]interface{}, len(bundle.Objects))
	for i, obj := range bundle.Objects {
		result[i] = obj
	}

	return result
}
