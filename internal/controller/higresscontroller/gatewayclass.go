package higresscontroller

import (
	"fmt"

	operatorv1alpha1 "github.com/alibaba/higress/higress-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gateway "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	group                = "gateway.networking.k8s.io"
	apiVersion           = "v1"
	kind                 = "GatewayClass"
	resource             = "gatewayclasses"
	gatewayClassResource = schema.GroupVersionResource{
		Group:    group,
		Version:  apiVersion,
		Resource: resource,
	}
)

func initGatewayclass(gatewayclass *gateway.GatewayClass, instance *operatorv1alpha1.HigressController) (*gateway.GatewayClass, error) {
	*gatewayclass = gateway.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Spec.IngressClass,
			Labels:      instance.Labels,
			Annotations: instance.Annotations,
		},
	}
	gatewayclass.APIVersion = fmt.Sprintf("%s/%s", group, apiVersion)
	gatewayclass.Kind = kind

	if _, err := updateGatewayclassSpec(gatewayclass, instance); err != nil {
		return nil, err
	}

	return gatewayclass, nil
}

func updateGatewayclassSpec(gatewayclass *gateway.GatewayClass, instance *operatorv1alpha1.HigressController) (*gateway.GatewayClass, error) {
	gatewayclass.Spec = gateway.GatewayClassSpec{
		//ControllerName: gateway.GatewayController("higress.io/" + instance.Spec.IngressClass),
		ControllerName: gateway.GatewayController("higress.io/gateway-controller"),
	}
	return gatewayclass, nil
}
