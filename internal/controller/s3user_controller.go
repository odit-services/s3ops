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
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/controller/s3client"
)

// S3UserReconciler reconciles a S3User object
type S3UserReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users/finalizers,verbs=update

func (r *S3UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.logger.Infow("Reconciling S3User", "name", req.Name, "namespace", req.Namespace)

	//TODO: Business Logic

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	var zapLogLevel zapcore.Level
	err := zapLogLevel.UnmarshalText([]byte(strings.ToLower(logLevel)))
	if err != nil {
		zapLogLevel = zapcore.InfoLevel
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapLogLevel)
	zapLogger, _ := zapConfig.Build()
	defer zapLogger.Sync()
	r.logger = zapLogger.Sugar()

	r.S3ClientFactory = &s3client.S3ClientFactoryDefault{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&s3oditservicesv1alpha1.S3User{}).
		Complete(r)
}
