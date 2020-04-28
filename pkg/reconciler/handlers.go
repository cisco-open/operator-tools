package reconciler

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ottypes "github.com/banzaicloud/operator-tools/pkg/types"
)

func EnqueueByOwnerAnnotationMapper() handler.Mapper {
	return handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
		var annotation string
		var ok bool
		if annotation, ok = a.Meta.GetAnnotations()[ottypes.BanzaiCloudRelatedTo]; !ok {
			return []reconcile.Request{}
		}
		pieces := strings.SplitN(annotation, string(types.Separator), 2)
		if len(pieces) != 2 {
			return []reconcile.Request{}
		}

		return []reconcile.Request{
			{NamespacedName: client.ObjectKey{
				Name:      pieces[1],
				Namespace: pieces[0],
			}},
		}
	})
}
