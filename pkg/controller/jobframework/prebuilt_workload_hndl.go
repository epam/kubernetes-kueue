package jobframework

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	"sigs.k8s.io/kueue/pkg/controller/constants"
	"sigs.k8s.io/kueue/pkg/util/api"
)

func PrebuiltWorkloadHndl[PL api.ObjListAsPtr[L], PE api.ObjAsPtr[E], L any, E any](c client.Client) handler.EventHandler {
	return &handler.Funcs{
		CreateFunc: func(ctx context.Context, tce event.CreateEvent, trli workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			QueueReconcileJobsWaitingForPrebuiltWorkload[PL, PE](ctx, c, tce.Object, trli)
		},
	}
}

func QueueReconcileJobsWaitingForPrebuiltWorkload[PL api.ObjListAsPtr[L], PE api.ObjAsPtr[E], L any, E any](ctx context.Context, c client.Client, object client.Object, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	w, ok := object.(*kueue.Workload)
	if !ok || len(w.OwnerReferences) > 0 {
		return
	}
	log := ctrl.LoggerFrom(ctx).WithValues("workload", klog.KObj(w))
	ctx = ctrl.LoggerInto(ctx, log)
	log.V(5).Info("Queueing reconcile for prebuilt workload waiting items")
	waitingJobs := PL(new(L))
	if err := c.List(ctx, waitingJobs, client.InNamespace(w.Namespace), client.MatchingLabels{constants.PrebuiltWorkloadLabel: w.Name}); err != nil {
		log.Error(err, "Unable to list waiting jobs")
		return
	}

	items, err := api.GetItemsFromList[E](waitingJobs)
	if err != nil {
		log.Error(err, "Unable to get list items")
		return
	}

	for i := range items {
		waitingJob := PE(&items[i])
		log.V(5).Info("Queueing reconcile for waiting items", "item", klog.KObj(waitingJob))
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      waitingJob.GetName(),
				Namespace: w.GetNamespace(),
			},
		})
	}
}
