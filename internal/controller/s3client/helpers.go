package s3client

import (
	"context"
	"fmt"
	"time"

	"github.com/odit-services/s3ops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getS3ServerObject(serverRef v1alpha1.ServerReference, r client.Client) (*v1alpha1.S3Server, metav1.Condition, error) {
	ctx := context.Background()

	s3Server := &v1alpha1.S3Server{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: serverRef.Namespace,
		Name:      serverRef.Name,
	}, s3Server)
	if err != nil {
		return nil, metav1.Condition{
			Type:               v1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             v1alpha1.ReasonNotFound,
			Message:            "S3Server resource not found",
			LastTransitionTime: metav1.Now(),
		}, err
	}

	return s3Server, metav1.Condition{}, nil
}

func GetS3ClientFromS3Server(serverRef v1alpha1.ServerReference, factory S3ClientFactory, r client.Client) (S3Client, metav1.Condition, error) {
	s3Server, condition, err := getS3ServerObject(serverRef, r)
	if err != nil {
		return nil, condition, err
	}

	s3Client, err := factory.NewClient(*s3Server)
	if err != nil {
		return nil, metav1.Condition{
			Type:    v1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  err.Error(),
			Message: fmt.Sprintf("Failed to create S3 client: %v", err),
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
		}, err
	}

	return s3Client, metav1.Condition{}, nil
}

func GetS3AdminClientFromS3Server(serverRef v1alpha1.ServerReference, factory S3ClientFactory, r client.Client) (S3AdminClient, metav1.Condition, error) {
	s3Server, condition, err := getS3ServerObject(serverRef, r)
	if err != nil {
		return nil, condition, err
	}

	s3AdminClient, err := factory.NewAdminClient(*s3Server)
	if err != nil {
		return nil, metav1.Condition{
			Type:    v1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  err.Error(),
			Message: fmt.Sprintf("Failed to create S3 admin client: %v", err),
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
		}, err
	}

	return s3AdminClient, metav1.Condition{}, nil
}
