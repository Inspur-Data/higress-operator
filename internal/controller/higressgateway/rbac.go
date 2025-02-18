package higressgateway

import (
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
)

const (
	role        = "higress-gateway"
	clusterRole = "higress-gateway"
)

func defaultRules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
	}

	return rules
}

func initClusterRole(cr *rbacv1.ClusterRole, instance *operatorv1alpha1.HigressGateway) *rbacv1.ClusterRole {
	*cr = rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Namespace + "-" + instance.Name,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		Rules: defaultRules(),
	}
	return cr
}

func muteClusterRole(cr *rbacv1.ClusterRole, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		cr.Rules = defaultRules()
		return nil
	}
}

func initClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressGateway) *rbacv1.ClusterRoleBinding {
	*crb = rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Namespace + "-" + instance.Name,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
	}

	updateClusterRoleBinding(crb, instance)
	return crb
}

func updateClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressGateway) {
	crb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     instance.Namespace + "-" + instance.Name,
		APIGroup: "rbac.authorization.k8s.io",
	}

	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      getServiceAccount(instance),
		Namespace: instance.Namespace,
	}

	for _, sub := range crb.Subjects {
		if reflect.DeepEqual(sub, subject) {
			return
		}
	}

	crb.Subjects = append(crb.Subjects, subject)
}

func muteClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		updateClusterRoleBinding(crb, instance)
		return nil
	}
}

func initRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressGateway) *rbacv1.RoleBinding {
	*rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     instance.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      getServiceAccount(instance),
				Namespace: instance.Namespace,
			},
		},
	}
	return rb
}

func updateRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressGateway) {
	rb.RoleRef = rbacv1.RoleRef{
		Kind:     "Role",
		Name:     instance.Name,
		APIGroup: "rbac.authorization.k8s.io",
	}

	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      getServiceAccount(instance),
		Namespace: instance.Namespace,
	}

	for _, sub := range rb.Subjects {
		if reflect.DeepEqual(sub, subject) {
			return
		}
	}

	rb.Subjects = append(rb.Subjects, subject)
}

func muteRoleBinding(rb *rbacv1.RoleBinding, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		updateRoleBinding(rb, instance)
		return nil
	}
}

func initRole(r *rbacv1.Role, instance *operatorv1alpha1.HigressGateway) *rbacv1.Role {
	*r = rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
		Rules: defaultRules(),
	}

	return r
}

func muteRole(role *rbacv1.Role, instance *operatorv1alpha1.HigressGateway) controllerutil.MutateFn {
	return func() error {
		role.Rules = defaultRules()
		return nil
	}
}
