// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package subnetmanager

import (
	"context"
	"errors"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/spidernet-io/spiderpool/pkg/constant"
	spiderpoolv1 "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v1"
	"github.com/spidernet-io/spiderpool/pkg/logutils"
)

var WebhookLogger *zap.Logger

type SubnetWebhook struct {
	client.Client

	EnableIPv4 bool
	EnableIPv6 bool
}

func (sw *SubnetWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if WebhookLogger == nil {
		WebhookLogger = logutils.Logger.Named("Subnet-Webhook")
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(&spiderpoolv1.SpiderSubnet{}).
		WithDefaulter(sw).
		WithValidator(sw).
		Complete()
}

var _ webhook.CustomDefaulter = (*SubnetWebhook)(nil)

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (sw *SubnetWebhook) Default(ctx context.Context, obj runtime.Object) error {
	subnet := obj.(*spiderpoolv1.SpiderSubnet)

	logger := WebhookLogger.Named("Mutating").With(
		zap.String("SubnetName", subnet.Name),
		zap.String("Operation", "DEFAULT"),
	)
	logger.Sugar().Debugf("Request Subnet: %+v", *subnet)

	if err := sw.mutateSubnet(logutils.IntoContext(ctx, logger), subnet); err != nil {
		logger.Sugar().Errorf("Failed to mutate Subnet: %v", err)
	}

	return nil
}

var _ webhook.CustomValidator = (*SubnetWebhook)(nil)

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (sw *SubnetWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	subnet := obj.(*spiderpoolv1.SpiderSubnet)

	logger := WebhookLogger.Named("Validating").With(
		zap.String("SubnetName", subnet.Name),
		zap.String("Operation", "CREATE"),
	)
	logger.Sugar().Debugf("Request Subnet: %+v", *subnet)

	if errs := sw.validateCreateSubnet(logutils.IntoContext(ctx, logger), subnet); len(errs) != 0 {
		logger.Sugar().Errorf("Failed to create Subnet: %v", errs.ToAggregate().Error())
		return apierrors.NewInvalid(
			schema.GroupKind{Group: constant.SpiderpoolAPIGroup, Kind: constant.SpiderSubnetKind},
			subnet.Name,
			errs,
		)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (sw *SubnetWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	oldSubnet := oldObj.(*spiderpoolv1.SpiderSubnet)
	newSubnet := newObj.(*spiderpoolv1.SpiderSubnet)

	logger := WebhookLogger.Named("Validating").With(
		zap.String("SubnetName", newSubnet.Name),
		zap.String("Operation", "UPDATE"),
	)
	logger.Sugar().Debugf("Request old Subnet: %+v", *oldSubnet)
	logger.Sugar().Debugf("Request new Subnet: %+v", *newSubnet)

	if newSubnet.DeletionTimestamp != nil {
		if !controllerutil.ContainsFinalizer(newSubnet, constant.SpiderFinalizer) {
			return nil
		}

		return apierrors.NewForbidden(
			schema.GroupResource{},
			"",
			errors.New("cannot update a terminaing Subnet"),
		)
	}

	if errs := sw.validateUpdateSubnet(logutils.IntoContext(ctx, logger), oldSubnet, newSubnet); len(errs) != 0 {
		logger.Sugar().Errorf("Failed to update Subnet: %v", errs.ToAggregate().Error())
		return apierrors.NewInvalid(
			schema.GroupKind{Group: constant.SpiderpoolAPIGroup, Kind: constant.SpiderSubnetKind},
			newSubnet.Name,
			errs,
		)
	}

	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (sw *SubnetWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}
