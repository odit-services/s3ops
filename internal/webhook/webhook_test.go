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

package webhook

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
)

var _ = Describe("Cross-Namespace Validation Webhook", func() {
	AfterEach(func() {
		os.Unsetenv("ENABLE_CROSS_NAMESPACE_VALIDATION")
	})

	Describe("validateServerRefCrossNamespace", func() {
		Context("when ENABLE_CROSS_NAMESPACE_VALIDATION is not set (default)", func() {
			BeforeEach(func() {
				os.Unsetenv("ENABLE_CROSS_NAMESPACE_VALIDATION")
			})

			It("should allow empty namespace in ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow ServerRef namespace matching resource namespace", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "default",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reject cross-namespace ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "other-namespace",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cross-namespace ServerRef is not allowed"))
				Expect(err.Error()).To(ContainSubstring("s3-server"))
				Expect(err.Error()).To(ContainSubstring("other-namespace"))
				Expect(err.Error()).To(ContainSubstring("default"))
			})
		})

		Context("when ENABLE_CROSS_NAMESPACE_VALIDATION is set to false", func() {
			BeforeEach(func() {
				os.Setenv("ENABLE_CROSS_NAMESPACE_VALIDATION", "false")
			})

			It("should allow empty namespace in ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow matching namespace", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "default",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reject cross-namespace ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "other-namespace",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when ENABLE_CROSS_NAMESPACE_VALIDATION is set to true", func() {
			BeforeEach(func() {
				os.Setenv("ENABLE_CROSS_NAMESPACE_VALIDATION", "true")
			})

			It("should allow empty namespace in ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow matching namespace", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "default",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow cross-namespace ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "other-namespace",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when ENABLE_CROSS_NAMESPACE_VALIDATION is set to 1", func() {
			BeforeEach(func() {
				os.Setenv("ENABLE_CROSS_NAMESPACE_VALIDATION", "1")
			})

			It("should allow cross-namespace ServerRef", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "other-namespace",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when ENABLE_CROSS_NAMESPACE_VALIDATION is set to invalid value", func() {
			BeforeEach(func() {
				os.Setenv("ENABLE_CROSS_NAMESPACE_VALIDATION", "invalid")
			})

			It("should default to disabled (reject cross-namespace)", func() {
				serverRef := s3oditservicesv1alpha1.ServerReference{
					Name:      "s3-server",
					Namespace: "other-namespace",
				}
				err := validateServerRefCrossNamespace(serverRef, "default")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})