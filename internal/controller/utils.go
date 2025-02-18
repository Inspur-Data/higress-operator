package controller

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdate(ctx context.Context, cli client.Client, kind string, object client.Object, f controllerutil.MutateFn, logger logr.Logger) error {
	key := client.ObjectKeyFromObject(object)
	status, err := controllerutil.CreateOrUpdate(ctx, cli, object, f)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate object {%s:%s}", kind, key))
		return err
	}

	logger.Info(fmt.Sprintf("createOrUpdate object {%s:%s} : %s", kind, key, status))
	return nil
}

func Delete(ctx context.Context, cli client.Client, kind string, object client.Object, logger logr.Logger) error {
	key := client.ObjectKeyFromObject(object)
	err := cli.Delete(ctx, object)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to delete object {%s:%s}", kind, key))
	}
	return nil
}

// CreateIfNotExits create obj if not exists
// true, nil : the obj exists
func CreateIfNotExits(ctx context.Context, cli client.Client, object client.Object) (bool, error) {
	var err error
	if err = cli.Create(ctx, object); err != nil && errors.IsAlreadyExists(err) {
		return true, nil
	}

	return false, err
}

func createOrUpdate(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn, logger logr.Logger) (controllerutil.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := mutate(f, key, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	existing := obj.DeepCopyObject()
	if err := mutate(f, key, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}

	if equality.Semantic.DeepEqual(existing, obj) {
		return controllerutil.OperationResultNone, nil
	}

	logger.Info(fmt.Sprintf("the diff of %v is %v", key, cmp.Diff(obj, existing)))

	if err := c.Update(ctx, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}
	return controllerutil.OperationResultUpdated, nil
}

func mutate(f controllerutil.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}
	return nil
}

func UpdateObjectMeta(obj *metav1.ObjectMeta, instance metav1.Object, labels map[string]string) {
	obj.Name = instance.GetName()
	obj.Namespace = instance.GetNamespace()
	obj.Labels = labels
}

func compair(olddeploy *appsv1.Deployment, newdeploy *appsv1.Deployment) bool {
	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Volumes, olddeploy.Spec.Template.Spec.Volumes) {
		return false
	}
	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Containers[1].VolumeMounts, olddeploy.Spec.Template.Spec.Containers[1].VolumeMounts) {
		return false
	}
	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Containers[1].Args, olddeploy.Spec.Template.Spec.Containers[1].Args) {
		return false
	}

	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Containers[1].Env, olddeploy.Spec.Template.Spec.Containers[1].Env) {
		return false
	}
	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Containers[1].Ports, olddeploy.Spec.Template.Spec.Containers[1].Ports) {
		return false
	}
	if !equality.Semantic.DeepEqual(newdeploy.Spec.Template.Spec.Containers[1].SecurityContext, olddeploy.Spec.Template.Spec.Containers[1].SecurityContext) {
		return false
	}
	return true
}

func CreateOrUpdateClusterResource(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, object client.Object, logger logr.Logger) error {
	name := object.GetName()
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate dynamicClient cluster object for {%s:%s}", gvr.Resource, name))
		return err
	}

	objectJson, err := json.Marshal(object)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate dynamicClient cluster object for {%s:%s}", gvr.Resource, name))
		return err
	}
	objectYaml := new(unstructured.Unstructured)
	err = json.Unmarshal(objectJson, objectYaml)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate dynamicClient cluster object {%s:%s}", gvr.Resource, name))
		return err
	}

	result, err := dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("Failed to get dynamicClient cluster object {%s:%s}", gvr.Resource, name))
			return err
		}
		_, err = dynamicClient.Resource(gvr).Create(ctx, objectYaml, metav1.CreateOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to create dynamicClient cluster object {%s:%s}", gvr.Resource, name))
			return err
		}
		logger.Info(fmt.Sprintf("create dynamicClient cluster object {%s:%s}", gvr.Resource, name))
		return nil
	}
	resultCopy := new(unstructured.Unstructured)
	resultCopy.Object = result.Object
	for key, value := range objectYaml.Object {
		resultCopy.Object[key] = value
	}
	if equality.Semantic.DeepEqual(result, resultCopy) {
		logger.Info(fmt.Sprintf("ignore to update dynamicClient cluster object {%s:%s}", gvr.Resource, name))
		return nil
	}
	_, err = dynamicClient.Resource(gvr).Update(ctx, resultCopy, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to update dynamicClient cluster object {%s:%s}", gvr.Resource, name))
		return err
	}
	logger.Info(fmt.Sprintf("update dynamicClient cluster object {%s:%s}", gvr.Resource, name))
	return nil
}

func DeleteClusterResource(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, object client.Object, logger logr.Logger) error {
	name := object.GetName()
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to createOrUpdate dynamicClient cluster object for {%s:%s}", gvr.Resource, name))
		return err
	}
	err = dynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("Failed to delete dynamicClient cluster object {%s:%s}", gvr.Resource, name))
			return err
		}
		logger.Info(fmt.Sprintf("ignore to delete dynamicClient cluster object {%s:%s}", gvr.Resource, name))
		return nil
	}
	logger.Info(fmt.Sprintf("delete dynamicClient cluster object {%s:%s}", gvr.Resource, name))
	return nil
}
