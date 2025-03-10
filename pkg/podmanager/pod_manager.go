// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package podmanager

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spidernet-io/spiderpool/pkg/constant"
	"github.com/spidernet-io/spiderpool/pkg/logutils"
	"github.com/spidernet-io/spiderpool/pkg/types"
)

type PodManager interface {
	GetPodByName(ctx context.Context, namespace, podName string) (*corev1.Pod, error)
	ListPods(ctx context.Context, opts ...client.ListOption) (*corev1.PodList, error)
	GetPodTopController(ctx context.Context, pod *corev1.Pod) (types.PodTopController, error)
}

type podManager struct {
	config PodManagerConfig
	client client.Client
}

func NewPodManager(config PodManagerConfig, client client.Client) (PodManager, error) {
	if client == nil {
		return nil, fmt.Errorf("k8s client %w", constant.ErrMissingRequiredParam)
	}

	return &podManager{
		config: setDefaultsForPodManagerConfig(config),
		client: client,
	}, nil
}

func (pm *podManager) GetPodByName(ctx context.Context, namespace, podName string) (*corev1.Pod, error) {
	var pod corev1.Pod
	if err := pm.client.Get(ctx, apitypes.NamespacedName{Namespace: namespace, Name: podName}, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

func (pm *podManager) ListPods(ctx context.Context, opts ...client.ListOption) (*corev1.PodList, error) {
	var podList corev1.PodList
	if err := pm.client.List(ctx, &podList, opts...); err != nil {
		return nil, err
	}

	return &podList, nil
}

// GetPodTopController will find the pod top owner controller with the given pod.
// For example, once we create a deployment then it will create replicaset and the replicaset will create pods.
// So, the pods' top owner is deployment. That's what the method implements.
// Notice: if the application is a third party controller, the types.PodTopController property App would be nil!
func (pm *podManager) GetPodTopController(ctx context.Context, pod *corev1.Pod) (types.PodTopController, error) {
	logger := logutils.FromContext(ctx)

	var ownerErr = fmt.Errorf("failed to get pod '%s/%s' owner", pod.Namespace, pod.Name)

	podOwner := metav1.GetControllerOf(pod)
	if podOwner == nil {
		return types.PodTopController{
			Kind:      constant.KindPod,
			Namespace: pod.Namespace,
			Name:      pod.Name,
			UID:       pod.UID,
			APP:       pod,
		}, nil
	}

	// third party controller
	if podOwner.APIVersion != appsv1.SchemeGroupVersion.String() && podOwner.APIVersion != batchv1.SchemeGroupVersion.String() {
		return types.PodTopController{
			Kind:      constant.KindUnknown,
			Namespace: pod.Namespace,
			Name:      podOwner.Name,
			UID:       podOwner.UID,
		}, nil
	}

	namespacedName := apitypes.NamespacedName{
		Namespace: pod.Namespace,
		Name:      podOwner.Name,
	}

	switch podOwner.Kind {
	case constant.KindReplicaSet:
		var replicaset appsv1.ReplicaSet
		err := pm.client.Get(ctx, namespacedName, &replicaset)
		if nil != err {
			return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
		}

		replicasetOwner := metav1.GetControllerOf(&replicaset)
		if replicasetOwner != nil {
			if replicasetOwner.Kind == constant.KindDeployment {
				var deployment appsv1.Deployment
				err = pm.client.Get(ctx, apitypes.NamespacedName{Namespace: replicaset.Namespace, Name: replicasetOwner.Name}, &deployment)
				if nil != err {
					return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
				}
				return types.PodTopController{
					Kind:      constant.KindDeployment,
					Namespace: deployment.Namespace,
					Name:      deployment.Name,
					UID:       deployment.UID,
					APP:       &deployment,
				}, nil
			}

			logger.Sugar().Warnf("the controller type '%s' of pod '%s/%s' is unknown", replicasetOwner.Kind, pod.Namespace, pod.Name)
			return types.PodTopController{
				Kind:      constant.KindUnknown,
				Namespace: pod.Namespace,
				Name:      replicasetOwner.Name,
				UID:       replicasetOwner.UID,
			}, nil
		}
		return types.PodTopController{
			Kind:      constant.KindReplicaSet,
			Namespace: replicaset.Namespace,
			Name:      replicaset.Name,
			UID:       replicaset.UID,
			APP:       &replicaset,
		}, nil

	case constant.KindJob:
		var job batchv1.Job
		err := pm.client.Get(ctx, namespacedName, &job)
		if nil != err {
			return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
		}
		jobOwner := metav1.GetControllerOf(&job)
		if jobOwner != nil {
			if jobOwner.Kind == constant.KindCronJob {
				var cronJob batchv1.CronJob
				err = pm.client.Get(ctx, apitypes.NamespacedName{Namespace: job.Namespace, Name: jobOwner.Name}, &cronJob)
				if nil != err {
					return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
				}
				return types.PodTopController{
					Kind:      constant.KindCronJob,
					Namespace: cronJob.Namespace,
					Name:      cronJob.Name,
					UID:       cronJob.UID,
					APP:       &cronJob,
				}, nil
			}

			logger.Sugar().Warnf("the controller type '%s' of pod '%s/%s' is unknown", jobOwner.Kind, pod.Namespace, pod.Name)
			return types.PodTopController{
				Kind:      constant.KindUnknown,
				Namespace: job.Namespace,
				Name:      jobOwner.Name,
				UID:       jobOwner.UID,
			}, nil
		}
		return types.PodTopController{
			Kind:      constant.KindJob,
			Namespace: job.Namespace,
			Name:      job.Name,
			UID:       job.UID,
			APP:       &job,
		}, nil

	case constant.KindDaemonSet:
		var daemonSet appsv1.DaemonSet
		err := pm.client.Get(ctx, namespacedName, &daemonSet)
		if nil != err {
			return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
		}
		return types.PodTopController{
			Kind:      constant.KindDaemonSet,
			Namespace: daemonSet.Namespace,
			Name:      daemonSet.Name,
			UID:       daemonSet.UID,
			APP:       &daemonSet,
		}, nil

	case constant.KindStatefulSet:
		var statefulSet appsv1.StatefulSet
		err := pm.client.Get(ctx, namespacedName, &statefulSet)
		if nil != err {
			return types.PodTopController{}, fmt.Errorf("%w: %v", ownerErr, err)
		}
		return types.PodTopController{
			Kind:      constant.KindStatefulSet,
			Namespace: statefulSet.Namespace,
			Name:      statefulSet.Name,
			UID:       statefulSet.UID,
			APP:       &statefulSet,
		}, nil
	}

	logger.Sugar().Warnf("the controller type '%s' of pod '%s/%s' is unknown", podOwner.Kind, pod.Namespace, pod.Name)
	return types.PodTopController{
		Kind:      constant.KindUnknown,
		Namespace: pod.Namespace,
		Name:      podOwner.Name,
		UID:       podOwner.UID,
	}, nil
}
