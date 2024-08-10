package mocks

import "github.com/odit-services/s3ops/api/v1alpha1"

func DefaultMockEnvs() S3ClientMockEnv {
	return S3ClientMockEnv{
		ValidEndpoints: []string{"valid.s3.odit.services", "valid2.s3.odit.services"},
		ValidCredentials: []v1alpha1.S3ServerAuthSpec{
			{
				AccessKey: "valid",
				SecretKey: "valid",
			},
		},
		ExistingBuckets: []string{"existing-bucket"},
	}
}
