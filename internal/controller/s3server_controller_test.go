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
	"github.com/odit-services/s3ops/internal/controller/mocks"
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
	s3MockEnv := mocks.DefaultMockEnvs()
	s3MockSpy := mocks.S3ClientMockSpy{}
	var testReconciler *S3ServerReconciler

	BeforeAll(func() {
		By("creating the test reconciler")
		testScheme := scheme.Scheme
		testScheme.AddKnownTypes(s3oditservicesv1alpha1.GroupVersion, &s3oditservicesv1alpha1.S3Server{})
		testReconciler = &S3ServerReconciler{
			Client: k8sClient,
			Scheme: testScheme,
			logger: zap.NewNop().Sugar(),
			S3ClientFactory: &mocks.S3ClientFactoryMocked{
				S3ClientMockEnv: &s3MockEnv,
				S3ClientMockSpy: &s3MockSpy,
			},
		}
	})

	Describe("Testing the reconcoile function", func() {
		Describe("Testing the reconciliation of a new s3server", func() {
			When("A new valid s3server is created", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				var result reconcile.Result
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
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
							Endpoint: s3MockEnv.ValidEndpoints[0],
							TLS:      true,
							Auth:     s3MockEnv.ValidCredentials[0],
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
				It("should set the status state to success", func() {
					Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateSuccess))
				})
				It("should set the status last reconcile time", func() {
					Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
				})
				It("should set the status condition to type online", func() {
					Expect(s3Server.Status.Online).To(BeTrue())
				})
			})
			When("A new invalid s3server is with an non-existant domain", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
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
				It("should set the status state to failed", func() {
					Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateFailed))
				})
				It("should set the status last reconcile time", func() {
					Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
				})
				It("should set the online status to false", func() {
					Expect(s3Server.Status.Online).To(BeFalse())
				})
			})
			When("A new invalid s3server is with an wrong credentials", func() {
				var err error
				var s3Server s3oditservicesv1alpha1.S3Server
				BeforeAll(func() {
					s3MockSpy = mocks.S3ClientMockSpy{}
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
							Endpoint: s3MockEnv.ValidEndpoints[0],
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
				It("should set the status state to failed", func() {
					Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateFailed))
				})
				It("should set the status last reconcile time", func() {
					Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
				})
				It("should set the online status to false", func() {
					Expect(s3Server.Status.Online).To(BeFalse())
				})
			})
		})
	})
	Describe("Testing the reconciliation of an existing updated s3server", func() {
		When("A valid s3server is updated with a new valid endpoint", func() {
			var err error
			var s3Server s3oditservicesv1alpha1.S3Server
			var result reconcile.Result
			BeforeAll(func() {
				s3MockSpy = mocks.S3ClientMockSpy{}
				nameSpacedName := types.NamespacedName{
					Name:      "test-s3server-update-endpoint",
					Namespace: "default",
				}
				s3Server = s3oditservicesv1alpha1.S3Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:      nameSpacedName.Name,
						Namespace: nameSpacedName.Namespace,
					},
					Spec: s3oditservicesv1alpha1.S3ServerSpec{
						Type:     "minio",
						Endpoint: s3MockEnv.ValidEndpoints[0],
						TLS:      true,
						Auth:     s3MockEnv.ValidCredentials[0],
					},
				}
				Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
				Expect(err).ToNot(HaveOccurred())

				testReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: nameSpacedName,
				})
				Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())

				s3Server.Spec.Endpoint = s3MockEnv.ValidEndpoints[1]
				Expect(k8sClient.Update(ctx, &s3Server)).To(Succeed())

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
			It("should set the status state to success", func() {
				Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateSuccess))
			})
			It("should set the status last reconcile time", func() {
				Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
			})
			It("should set the online status to true", func() {
				Expect(s3Server.Status.Online).To(BeTrue())
			})
		})

		When("An invalid s3server is updated with a valid endpoint", func() {
			var err error
			var s3Server s3oditservicesv1alpha1.S3Server
			var result reconcile.Result
			BeforeAll(func() {
				s3MockSpy = mocks.S3ClientMockSpy{}
				nameSpacedName := types.NamespacedName{
					Name:      "test-invalid-s3server-update-endpoint",
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
						Auth:     s3MockEnv.ValidCredentials[0],
					},
				}
				Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
				Expect(err).ToNot(HaveOccurred())

				testReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: nameSpacedName,
				})
				Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())

				s3Server.Spec.Endpoint = s3MockEnv.ValidEndpoints[0]
				Expect(k8sClient.Update(ctx, &s3Server)).To(Succeed())

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
			It("should set the status state to success", func() {
				Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateSuccess))
			})
			It("should set the status last reconcile time", func() {
				Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
			})
			It("should set the online status to true", func() {
				Expect(s3Server.Status.Online).To(BeTrue())
			})
		})
		When("A valid s3server is updated with a new invalid endpoint", func() {
			var err error
			var s3Server s3oditservicesv1alpha1.S3Server
			BeforeAll(func() {
				s3MockSpy = mocks.S3ClientMockSpy{}
				nameSpacedName := types.NamespacedName{
					Name:      "test-s3server-update-invalid-endpoint",
					Namespace: "default",
				}
				s3Server = s3oditservicesv1alpha1.S3Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:      nameSpacedName.Name,
						Namespace: nameSpacedName.Namespace,
					},
					Spec: s3oditservicesv1alpha1.S3ServerSpec{
						Type:     "minio",
						Endpoint: s3MockEnv.ValidEndpoints[0],
						TLS:      true,
						Auth:     s3MockEnv.ValidCredentials[0],
					},
				}
				Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
				Expect(err).ToNot(HaveOccurred())

				_, err = testReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: nameSpacedName,
				})
				Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())

				s3Server.Spec.Endpoint = "invalid-domain.s3.odit.services"
				Expect(k8sClient.Update(ctx, &s3Server)).To(Succeed())

				_, err = testReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: nameSpacedName,
				})
				Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("should set the status state to failed", func() {
				Expect(s3Server.Status.State).To(Equal(s3oditservicesv1alpha1.StateFailed))
			})
			It("should set the status last reconcile time", func() {
				Expect(s3Server.Status.LastReconcileTime).ToNot(BeEmpty())
			})
			It("should set the online status to false", func() {
				Expect(s3Server.Status.Online).To(BeFalse())
			})
		})
	})
	Describe("Testing the reconciliation of an existing deleted s3server", func() {
		When("A valid s3server is deleted", func() {
			var err error
			var s3Server s3oditservicesv1alpha1.S3Server
			BeforeAll(func() {
				s3MockSpy = mocks.S3ClientMockSpy{}
				nameSpacedName := types.NamespacedName{
					Name:      "test-s3server-delete",
					Namespace: "default",
				}
				s3Server = s3oditservicesv1alpha1.S3Server{
					ObjectMeta: metav1.ObjectMeta{
						Name:      nameSpacedName.Name,
						Namespace: nameSpacedName.Namespace,
					},
					Spec: s3oditservicesv1alpha1.S3ServerSpec{
						Type:     "minio",
						Endpoint: s3MockEnv.ValidEndpoints[0],
						TLS:      true,
						Auth:     s3MockEnv.ValidCredentials[0],
					},
				}
				Expect(k8sClient.Create(ctx, &s3Server)).To(Succeed())
				Expect(err).ToNot(HaveOccurred())

				_, err = testReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: nameSpacedName,
				})
				Expect(k8sClient.Get(ctx, nameSpacedName, &s3Server)).To(Succeed())

				Expect(k8sClient.Delete(ctx, &s3Server)).To(Succeed())
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should no longer exist in k8s", func() {
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      s3Server.Name,
					Namespace: s3Server.Namespace,
				}, &s3Server)).To(HaveOccurred())
			})
		})
	})
})
