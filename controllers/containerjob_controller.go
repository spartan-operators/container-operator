/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	containerv1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
)

// ContainerJobReconciler reconciles a ContainerJob object
type ContainerJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=container.vnv.wrstudio.cloud,resources=containerjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=container.vnv.wrstudio.cloud,resources=containerjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=container.vnv.wrstudio.cloud,resources=containerjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=*,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ContainerJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ContainerJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Request Namespaced name: ", "namespace.name.request", req.NamespacedName)

	// Fetch the VrTestJob Resource
	containerJob := &containerv1alpha1.ContainerJob{}
	err := r.Get(ctx, req.NamespacedName, containerJob)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ContainerJob resource not found. Ignoring since the object will be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Error while fetching ContainerJob resource")
	}

	// Check if the Job already exists, If not create a new one
	found := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: containerJob.ObjectMeta.Namespace,
		Name:      fmt.Sprintf("%s-%s", containerJob.Name, containerJob.Spec.JobId),
	}, found)

	log.Info("Input Information: ", "******", containerJob.ObjectMeta.String())
	log.Info("Find Information: ", "******", found.String())

	if err != nil && errors.IsNotFound(err) {
		job := containerJobDescriptor(containerJob)
		log.Info("****** Creating a new Job ******", "Job.Namespace", job.Namespace, "Job.Name", job.Name, "Job.Image", job.Spec.Template.Spec.Containers)

		if err = r.Create(ctx, job); err != nil {
			log.Error(err, "****** Failed to create new Job ******", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		//update CR status with the pod names
		updateCrOutputs(&log, containerJob, r, &ctx)
		// Job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil

	} else if err != nil && errors.IsAlreadyExists(err) {
		log.Info("****** Requested Job already exists ******", "Job.Namespace", found.Namespace, "Job.Name", found.Name, "Job.Status", found.Status.String())

		return ctrl.Result{Requeue: true}, nil
	} else if jobStatus(&log, found) {
		//update CR status with the pod names
		updateCrOutputs(&log, containerJob, r, &ctx)

		// Job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil

	} else {
		log.Error(err, "****** Failed to get the Job Spec ******")
		return ctrl.Result{}, err
	}
}

func containerJobDescriptor(containerJob *containerv1alpha1.ContainerJob) *batchv1.Job {
	jobName := fmt.Sprintf("%s-%s", containerJob.Name, containerJob.Spec.JobId)
	containerName := fmt.Sprintf("%s-container-%s", containerJob.Name, containerJob.Spec.JobId)
	containerImage := fmt.Sprintf("%s:%s", containerJob.Spec.ImageSpec.Name, containerJob.Spec.ImageSpec.Version)
	containerArgs := appendArgs(containerJob)

	metadata := &metav1.ObjectMeta{
		Name:        jobName,
		Namespace:   containerJob.Namespace,
		Labels:      getContainerJobLabels(containerJob),
		Annotations: map[string]string{},
	}

	podSpec := &apiv1.PodSpec{
		Containers: []apiv1.Container{
			{
				Name:    containerName,
				Image:   containerImage,
				Command: []string{},
				Args:    containerArgs,
			},
		},
		RestartPolicy: apiv1.RestartPolicyOnFailure,
	}

	podTemplate := &apiv1.PodTemplateSpec{
		ObjectMeta: *metadata,
		Spec:       *podSpec,
	}

	jobSpec := &batchv1.JobSpec{
		Template: *podTemplate,
	}

	return &batchv1.Job{
		ObjectMeta: *metadata,
		Spec:       *jobSpec,
	}
}

func appendArgs(containerJob *containerv1alpha1.ContainerJob) []string {
	specArgs := containerJob.Spec.ParameterSpec.Args
	args := make([]string, 0)

	if len(specArgs) > 0 {
		args = append(args, specArgs)
	}

	return args
}

func jobStatus(log *logr.Logger, jobSpec *batchv1.Job) bool {
	var result = false

	if jobSpec.Status.Conditions == nil || len(jobSpec.Status.Conditions) < 0 {
		(*log).Info("****** Job Status Conditions are empty... no information can be retrieved ******", "Job.Namespace", jobSpec.Namespace, "Job.Name", jobSpec.Name, "Job.Status", jobSpec.Status.String())
		return true
	}

	for _, condition := range jobSpec.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			(*log).Info("****** Requested Job already exists and it was completed ******", "Job.Namespace", jobSpec.Namespace, "Job.Name", jobSpec.Name, "Job.Status", jobSpec.Status.String())
			result = true
		} else if condition.Type == batchv1.JobFailed {
			(*log).Info("****** Requested Job already exists and it was failed ******", "Job.Namespace", jobSpec.Namespace, "Job.Name", jobSpec.Name, "Job.Status", jobSpec.Status.String())
			result = true
		}
		break
	}

	return result
}

func updateCrOutputs(log *logr.Logger, containerJob *containerv1alpha1.ContainerJob, r *ContainerJobReconciler, ctx *context.Context) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(containerJob.Namespace),
		client.MatchingLabels(getContainerJobLabels(containerJob)),
	}
	if err := r.List(*ctx, podList, listOpts...); err != nil {
		(*log).Error(err, "Failed to list pods", "Job.Namespace", containerJob.Namespace, "Job.Name", containerJob.Name)
	}
	podNames := getPodNames(podList.Items)

	(*log).Info("****** Pods Retrieved ******", "Job.Pods", podNames)

	if podNames != nil || len(podNames) > 0 {
		// Update status.Nodes
		(*containerJob).Status.Nodes = podNames
		(*log).Info("****** ContainerJob Updated  ******", "CR.content", *containerJob)
		if err := r.Status().Update(*ctx, containerJob); err != nil {
			(*log).Error(err, "Failed to update Job status")
		}

	} else {
		(*log).Info("****** ContainerJob NOT UPDATED since pod names are empty ******")
	}
}

func getContainerJobLabels(containerJob *containerv1alpha1.ContainerJob) map[string]string {
	return map[string]string{
		"job_name": fmt.Sprintf("%s-%s", containerJob.Name, containerJob.Spec.JobId),
	}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerv1alpha1.ContainerJob{}).
		Complete(r)
}
