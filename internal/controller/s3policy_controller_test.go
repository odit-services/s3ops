/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	"github.com/odit-services/s3ops/internal/controller/mocks"
)

var _ = Describe("S3Policy Controller", Ordered, func() {
	ctx := context.Background()
	s3MockEnv := mocks.DefaultMockEnvs()
	s3MockSpy := mocks.S3ClientMockSpy{}
	var s3Server *s3oditservicesv1alpha1.S3Server
	var s3ServerBroken *s3oditservicesv1alpha1.S3Server
	var testReconciler *S3PolicyReconciler

	BeforeAll(func() {
		By("creating the test reconciler")
		testScheme := scheme.Scheme
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3Server{})
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3User{})
		testReconciler = &S3PolicyReconciler{
			Client: k8sClient,
			Scheme: testScheme,
			logger: zap.NewNop().Sugar(),
			S3ClientFactory: &mocks.S3ClientFactoryMocked{
				S3ClientMockEnv: &s3MockEnv,
				S3ClientMockSpy: &s3MockSpy,
			},
		}
		serverReconciler := &S3ServerReconciler{
			Client: k8sClient,
			Scheme: testScheme,
			logger: zap.NewNop().Sugar(),
			S3ClientFactory: &mocks.S3ClientFactoryMocked{
				S3ClientMockEnv: &s3MockEnv,
				S3ClientMockSpy: &s3MockSpy,
			},
		}
		By("creating a test s3 server")
		s3Server = &s3oditservicesv1alpha1.S3Server{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-s3-server-policy",
				Namespace: "default",
			},
			Spec: s3oditservicesv1alpha1.S3ServerSpec{
				Type:     "minio",
				Endpoint: s3MockEnv.ValidEndpoints[0],
				TLS:      true,
				Auth:     s3MockEnv.ValidCredentials[0],
			},
		}
		Expect(k8sClient.Create(ctx, s3Server)).To(Succeed())
		serverReconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      s3Server.Name,
				Namespace: s3Server.Namespace,
			},
		})

		By("creating a invalid test s3 server")
		s3ServerBroken = &s3oditservicesv1alpha1.S3Server{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-s3-server-invalid-policy",
				Namespace: "default",
			},
			Spec: s3oditservicesv1alpha1.S3ServerSpec{
				Type:     "minio",
				Endpoint: s3MockEnv.ValidEndpoints[0],
				TLS:      true,
				Auth: s3oditservicesv1alpha1.S3ServerAuthSpec{
					AccessKey: "invalid",
					SecretKey: "invalid",
				},
			},
		}
		Expect(k8sClient.Create(ctx, s3ServerBroken)).To(Succeed())
		serverReconciler.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      s3ServerBroken.Name,
				Namespace: s3ServerBroken.Namespace,
			},
		})
	})

	Describe("Testing the reconcoile function", func() {
		Describe("Testing the reconciliation of a new s3policy", func() {
			When("A new valid s3policy is created with a valid s3server", func() {
				var err error
				var result ctrl.Result
				var s3Policy s3oditservicesv1alpha1.S3Policy
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3-policy-nonexistent",
						Namespace: "default",
					}
					s3Policy = s3oditservicesv1alpha1.S3Policy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3PolicySpec{
							ServerRef: s3oditservicesv1alpha1.ServerReference{
								Name:      s3Server.Name,
								Namespace: s3Server.Namespace,
							},
							PolicyContent: fmt.Sprintf(`
								{
									"Version": "2012-10-17",
									"Statement": [
										{
											"Sid": "TestS3Policy",
											"Effect": "Allow",
											"Action": [
												"s3:ListBucket",
												"s3:PutObject",
												"s3:GetObject"
											],
											"Resource": [
												"arn:aws:s3:::%s",
												"arn:aws:s3:::%s/*"
											]
										}
									]
								}
							`, s3MockEnv.ExistingBuckets[0], s3MockEnv.ExistingBuckets[0]),
						},
					}
					Expect(k8sClient.Create(ctx, &s3Policy)).To(Succeed())

					result, err = testReconciler.Reconcile(ctx, ctrl.Request{
						NamespacedName: nameSpacedName,
					})
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Policy)).To(Succeed())
				})
				It("Should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
				It("should return a result with requeue set higher than 0", func() {
					Expect(result.RequeueAfter).To(BeNumerically(">", 0))
				})
				It("should set the status state to success", func() {
					Expect(s3Policy.Status.State).To(Equal(s3oditservicesv1alpha1.StateSuccess))
				})
				It("should set the status last reconcile time", func() {
					Expect(s3Policy.Status.LastReconcileTime).ToNot(BeEmpty())
				})
				It("should call the s3client exists function once", func() {
					Expect(s3MockSpy.PolicyExistsCalled).To(Equal(1))
				})
				It("should call the s3client make function once", func() {
					Expect(s3MockSpy.MakePolicyCalled).To(Equal(1))
				})
				It("Should set the status name field to a name matching the generation spec", func() {
					log.Printf("Status name: %s", s3Policy.Status.Name)
					log.Printf("Status: %v", s3Policy.Status)
					Expect(s3Policy.Status.Name).To(MatchRegexp("test-s3-policy-nonexistent-default-.+"))
				})
			})
		})
		Describe("Testing the reconciliation of a existing s3policy", func() {
			When("A valid s3policy is updated with new policy content with a valid s3server", func() {
				var err error
				var result ctrl.Result
				var s3Policy s3oditservicesv1alpha1.S3Policy
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3-policy-update",
						Namespace: "default",
					}
					s3Policy = s3oditservicesv1alpha1.S3Policy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3PolicySpec{
							ServerRef: s3oditservicesv1alpha1.ServerReference{
								Name:      s3Server.Name,
								Namespace: s3Server.Namespace,
							},
							PolicyContent: fmt.Sprintf(`
								{
									"Version": "2012-10-17",
									"Statement": [
										{
											"Sid": "TestS3Policy",
											"Effect": "Allow",
											"Action": [
												"s3:ListBucket",
												"s3:PutObject",
												"s3:GetObject"
											],
											"Resource": [
												"arn:aws:s3:::%s",
												"arn:aws:s3:::%s/*"
											]
										}
									]
								}
							`, s3MockEnv.ExistingBuckets[0], s3MockEnv.ExistingBuckets[0]),
						},
					}
					Expect(k8sClient.Create(ctx, &s3Policy)).To(Succeed())
					_, err = testReconciler.Reconcile(ctx, ctrl.Request{
						NamespacedName: nameSpacedName,
					})
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Policy)).To(Succeed())

					s3MockSpy = mocks.S3ClientMockSpy{}
					s3MockEnv.ExistingPolicies = append(s3MockEnv.ExistingPolicies, s3Policy.Status.Name)
					s3Policy.Spec.PolicyContent = fmt.Sprintf(`
						{
							"Version": "2012-10-17",
							"Statement": [
								{
									"Sid": "TestS3Policy",
									"Effect": "Allow",
									"Action": [
										"s3:ListBucket",
										"s3:PutObject",
										"s3:GetObject"
									],
									"Resource": [
										"arn:aws:s3:::%s",
										"arn:aws:s3:::%s/*"
									]
								}
							]
						}
					`, s3MockEnv.ExistingBuckets[0], s3MockEnv.ExistingBuckets[0])
					Expect(k8sClient.Update(ctx, &s3Policy)).To(Succeed())
					result, err = testReconciler.Reconcile(ctx, ctrl.Request{
						NamespacedName: nameSpacedName,
					})
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Policy)).To(Succeed())
				})
				It("Should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
				It("should return a result with requeue set higher than 0", func() {
					Expect(result.RequeueAfter).To(BeNumerically(">", 0))
				})
				It("should set the status state to success", func() {
					Expect(s3Policy.Status.State).To(Equal(s3oditservicesv1alpha1.StateSuccess))
				})
				It("should set the status last reconcile time", func() {
					Expect(s3Policy.Status.LastReconcileTime).ToNot(BeEmpty())
				})
				It("should call the s3client exists function once", func() {
					Expect(s3MockSpy.PolicyExistsCalled).To(Equal(1))
				})
				It("should call the s3client update function once", func() {
					Expect(s3MockSpy.UpdatePolicyCalled).To(Equal(1))
				})
			})
		})
		Describe("Testing the reconciliation of a deleted s3policy", func() {
			When("A new valid s3policy is deleted", func() {
				var err error
				var s3Policy s3oditservicesv1alpha1.S3Policy
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3-policy-deleteme",
						Namespace: "default",
					}
					s3Policy = s3oditservicesv1alpha1.S3Policy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3PolicySpec{
							ServerRef: s3oditservicesv1alpha1.ServerReference{
								Name:      s3Server.Name,
								Namespace: s3Server.Namespace,
							},
							PolicyContent: fmt.Sprintf(`
								{
									"Version": "2012-10-17",
									"Statement": [
										{
											"Sid": "TestS3Policy",
											"Effect": "Allow",
											"Action": [
												"s3:ListBucket",
												"s3:PutObject",
												"s3:GetObject"
											],
											"Resource": [
												"arn:aws:s3:::%s",
												"arn:aws:s3:::%s/*"
											]
										}
									]
								}
							`, s3MockEnv.ExistingBuckets[0], s3MockEnv.ExistingBuckets[0]),
						},
					}
					Expect(k8sClient.Create(ctx, &s3Policy)).To(Succeed())

					_, err = testReconciler.Reconcile(ctx, ctrl.Request{
						NamespacedName: nameSpacedName,
					})
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Policy)).To(Succeed())
					s3MockSpy = mocks.S3ClientMockSpy{}
					s3MockEnv.ExistingPolicies = append(s3MockEnv.ExistingPolicies, s3Policy.Status.Name)

					Expect(k8sClient.Delete(ctx, &s3Policy)).To(Succeed())
					_, err = testReconciler.Reconcile(ctx, ctrl.Request{
						NamespacedName: nameSpacedName,
					})

				})
				It("Should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
				It("should have deleted the s3policy ressource", func() {
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      s3Policy.Name,
						Namespace: s3Policy.Namespace,
					}, &s3Policy)).ToNot(Succeed())
				})
				It("should call the s3client policy exists function once", func() {
					Expect(s3MockSpy.PolicyExistsCalled).To(Equal(1))
				})
				It("should call the s3client remove policy function once", func() {
					Expect(s3MockSpy.RemovePolicyCalled).To(Equal(1))
				})
			})
		})
	})
})
