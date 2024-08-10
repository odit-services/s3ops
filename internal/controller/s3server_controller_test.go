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
	"log"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("S3Server Controller", Ordered, func() {
	ctx := context.Background()
	var testReconciler *S3ServerReconciler

	BeforeAll(func() {
		By("creating the test reconciler")
		testScheme := scheme.Scheme
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3Server{})
		testReconciler = &S3ServerReconciler{
			Client: k8sClient,
			Scheme: testScheme,
			logger: zap.NewNop().Sugar(),
		}
	})

	Describe("Testing the reconcoile function", func() {
		Describe("Testing the reconciliation of a new s3server", func() {
			When("A new valid s3server is created", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				var result reconcile.Result
				BeforeAll(func() {
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3server",
						Namespace: "default",
					}
					s3Server = s3oditservicesv1alpha1.S3Server{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3ServerSpec{
							Type:     "minio",
							Endpoint: "play.min.io",
							TLS:      true,
							Auth: s3oditservicesv1alpha1.S3ServerAuthSpec{
								AccessKey: "Q3AM3UQ867SPQQA43P2F",
								SecretKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
							},
						},
					}
					Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
					Expect(err).ToNot(HaveOccurred())

					result, err = testReconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: nameSpacedName,
					})
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
				It("should return a result with requeue set higher than 0", func() {
					Expect(result.RequeueAfter).To(BeNumerically(">", 0))
				})
				It("should set the status condition to type ready", func() {
					Expect(s3Server.Status.Conditions[len(s3Server.Status.Conditions)-1].Type).To(Equal(s3oditservicesv1alpha1.ConditionReady))
				})
			})
			When("A new invalid s3server is with an non-existant domain", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				BeforeAll(func() {
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3server-invalid-domain",
						Namespace: "default",
					}
					s3Server = s3oditservicesv1alpha1.S3Server{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3ServerSpec{
							Type:     "minio",
							Endpoint: "invalid-domain.s3.odit.services",
							TLS:      true,
							Auth: s3oditservicesv1alpha1.S3ServerAuthSpec{
								AccessKey: "Q3AM3UQ867SPQQA43P2F",
								SecretKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
							},
						},
					}
					Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
					Expect(err).ToNot(HaveOccurred())

					_, err = testReconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: nameSpacedName,
					})
					log.Printf("Error: %v", err)
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
				It("should set the status condition to type failed", func() {
					Expect(s3Server.Status.Conditions[len(s3Server.Status.Conditions)-1].Type).To(Equal(s3oditservicesv1alpha1.ConditionFailed))
				})
			})
			When("A new invalid s3server is with an wrong credentials", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				BeforeAll(func() {
					nameSpacedName := types.NamespacedName{
						Name:      "test-s3server-invalid-credentials",
						Namespace: "default",
					}
					s3Server = s3oditservicesv1alpha1.S3Server{
						ObjectMeta: metav1.ObjectMeta{
							Name:      nameSpacedName.Name,
							Namespace: nameSpacedName.Namespace,
						},
						Spec: s3oditservicesv1alpha1.S3ServerSpec{
							Type:     "minio",
							Endpoint: "play.min.io",
							TLS:      true,
							Auth: s3oditservicesv1alpha1.S3ServerAuthSpec{
								AccessKey: "invalid",
								SecretKey: "invalid",
							},
						},
					}
					Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
					Expect(err).ToNot(HaveOccurred())

					_, err = testReconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: nameSpacedName,
					})
					log.Printf("Error: %v", err)
					Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
				It("should set the status condition to type failed", func() {
					Expect(s3Server.Status.Conditions[len(s3Server.Status.Conditions)-1].Type).To(Equal(s3oditservicesv1alpha1.ConditionFailed))
				})
			})
		})
	})
})
