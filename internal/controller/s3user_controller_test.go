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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	"github.com/odit-services/s3ops/internal/controller/mocks"
)

var _ = Describe("S3User Controller", Ordered, func() {
	ctx := context.Background()
	s3MockEnv := mocks.DefaultMockEnvs()
	s3MockSpy := mocks.S3ClientMockSpy{}
	var s3Server *s3oditservicesv1alpha1.S3Server
	var s3ServerBroken *s3oditservicesv1alpha1.S3Server
	var testReconciler *S3UserReconciler

	BeforeAll(func() {
		By("creating the test reconciler")
		testScheme := scheme.Scheme
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3Server{})
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3User{})
		testReconciler = &S3UserReconciler{
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
				Name:      "test-s3-server",
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

		By("creating a invalid test s3 server")
		s3ServerBroken = &s3oditservicesv1alpha1.S3Server{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-s3-server-invalid",
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
		Describe("Testing the reconcoile function", func() {
			Describe("Testing the reconciliation of a new s3user", func() {
				When("A new valid s3bucket is created with a valid s3server", func() {
					var err error
					var result ctrl.Result
					var s3User s3oditservicesv1alpha1.S3User
					BeforeAll(func() {
						s3MockSpy = mocks.S3ClientMockSpy{}
						nameSpacedName := types.NamespacedName{
							Name:      "test-s3-user-nonexistent",
							Namespace: "default",
						}
						s3User = s3oditservicesv1alpha1.S3User{
							ObjectMeta: metav1.ObjectMeta{
								Name:      nameSpacedName.Name,
								Namespace: nameSpacedName.Namespace,
							},
							Spec: s3oditservicesv1alpha1.S3UserSpec{
								ServerRef: s3oditservicesv1alpha1.ServerReference{
									Name:      s3Server.Name,
									Namespace: s3Server.Namespace,
								},
							},
						}
						Expect(k8sClient.Create(ctx, &s3User)).To(Succeed())

						result, err = testReconciler.Reconcile(ctx, ctrl.Request{
							NamespacedName: nameSpacedName,
						})
						Expect(k8sClient.Get(ctx, nameSpacedName, &s3User)).To(Succeed())
					})

					It("Should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
					It("should return a result with requeue set higher than 0", func() {
						Expect(result.RequeueAfter).To(BeNumerically(">", 0))
					})
					It("should set the status condition to type ready", func() {
						Expect(s3User.Status.Conditions[len(s3User.Status.Conditions)-1].Type).To(Equal(s3oditservicesv1alpha1.ConditionReady))
					})
					It("Should set the status created to true", func() {
						Expect(s3User.Status.Created).To(BeTrue())
					})
					It("Should set the status secret ref to a valid secret", func() {
						Expect(s3User.Status.SecretRef).ToNot(BeEmpty())
					})
					It("Should create a secret in the same namespace as the user", func() {
						secret := &corev1.Secret{}
						Expect(k8sClient.Get(ctx, types.NamespacedName{
							Name:      s3User.Status.SecretRef,
							Namespace: s3User.Namespace,
						}, secret)).To(Succeed())
					})
					It("Should call the user exists function once", func() {
						Expect(s3MockSpy.UserExistsCalled).To(Equal(1))
					})
					It("Should call the make user function once", func() {
						Expect(s3MockSpy.MakeUserCalled).To(Equal(1))
					})
				})
			})
			Describe("Testing the reconciliation of a existing s3user", func() {
				When("A valid s3 user get's updated without any changes", func() {
					var err error
					var result ctrl.Result
					var s3User s3oditservicesv1alpha1.S3User
					BeforeAll(func() {
						s3MockSpy = mocks.S3ClientMockSpy{}
						nameSpacedName := types.NamespacedName{
							Name:      "test-s3-user-update-nochanges",
							Namespace: "default",
						}
						s3User = s3oditservicesv1alpha1.S3User{
							ObjectMeta: metav1.ObjectMeta{
								Name:      nameSpacedName.Name,
								Namespace: nameSpacedName.Namespace,
							},
							Spec: s3oditservicesv1alpha1.S3UserSpec{
								ServerRef: s3oditservicesv1alpha1.ServerReference{
									Name:      s3Server.Name,
									Namespace: s3Server.Namespace,
								},
							},
						}
						Expect(k8sClient.Create(ctx, &s3User)).To(Succeed())

						result, err = testReconciler.Reconcile(ctx, ctrl.Request{
							NamespacedName: nameSpacedName,
						})
						s3MockSpy = mocks.S3ClientMockSpy{}
						result, err = testReconciler.Reconcile(ctx, ctrl.Request{
							NamespacedName: nameSpacedName,
						})
						Expect(k8sClient.Get(ctx, nameSpacedName, &s3User)).To(Succeed())
					})

					It("Should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
					It("should return a result with requeue set higher than 0", func() {
						Expect(result.RequeueAfter).To(BeNumerically(">", 0))
					})
					It("should set the status condition to type ready", func() {
						Expect(s3User.Status.Conditions[len(s3User.Status.Conditions)-1].Type).To(Equal(s3oditservicesv1alpha1.ConditionReady))
					})
					It("Should set the status secret ref to a valid secret", func() {
						Expect(s3User.Status.SecretRef).ToNot(BeEmpty())
					})
					It("Should create a secret in the same namespace as the user", func() {
						secret := &corev1.Secret{}
						Expect(k8sClient.Get(ctx, types.NamespacedName{
							Name:      s3User.Status.SecretRef,
							Namespace: s3User.Namespace,
						}, secret)).To(Succeed())
					})
					It("Should call the user exists function once", func() {
						Expect(s3MockSpy.UserExistsCalled).To(Equal(1))
					})
					It("Should never call the make user function", func() {
						Expect(s3MockSpy.MakeUserCalled).To(Equal(0))
					})
				})
			})
		})
	})
})
