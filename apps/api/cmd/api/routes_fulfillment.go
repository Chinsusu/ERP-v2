package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type fulfillmentRouteHandlers struct {
	pickTasks                     routeHandler
	pickTaskDetail                routeHandler
	pickTaskStart                 routeHandler
	pickTaskConfirmLine           routeHandler
	pickTaskComplete              routeHandler
	pickTaskException             routeHandler
	packTasks                     routeHandler
	packTaskDetail                routeHandler
	packTaskStart                 routeHandler
	packTaskConfirm               routeHandler
	packTaskException             routeHandler
	carrierManifests              routeHandler
	carrierManifestAddShipment    routeHandler
	carrierManifestRemoveShipment routeHandler
	carrierManifestReady          routeHandler
	carrierManifestCancel         routeHandler
	carrierManifestExceptions     routeHandler
	carrierManifestHandover       routeHandler
	carrierManifestScan           routeHandler
}

func registerFulfillmentRoutes(routes routeGroup, handlers fulfillmentRouteHandlers) {
	routes.token("/api/v1/pick-tasks", handlers.pickTasks)
	routes.token("/api/v1/pick-tasks/{pick_task_id}", handlers.pickTaskDetail)
	routes.permission("/api/v1/pick-tasks/{pick_task_id}/start", auth.PermissionRecordCreate, handlers.pickTaskStart)
	routes.permission("/api/v1/pick-tasks/{pick_task_id}/confirm-line", auth.PermissionRecordCreate, handlers.pickTaskConfirmLine)
	routes.permission("/api/v1/pick-tasks/{pick_task_id}/complete", auth.PermissionRecordCreate, handlers.pickTaskComplete)
	routes.permission("/api/v1/pick-tasks/{pick_task_id}/exception", auth.PermissionRecordCreate, handlers.pickTaskException)
	routes.token("/api/v1/pack-tasks", handlers.packTasks)
	routes.token("/api/v1/pack-tasks/{pack_task_id}", handlers.packTaskDetail)
	routes.permission("/api/v1/pack-tasks/{pack_task_id}/start", auth.PermissionRecordCreate, handlers.packTaskStart)
	routes.permission("/api/v1/pack-tasks/{pack_task_id}/confirm", auth.PermissionRecordCreate, handlers.packTaskConfirm)
	routes.permission("/api/v1/pack-tasks/{pack_task_id}/exception", auth.PermissionRecordCreate, handlers.packTaskException)
	routes.token("/api/v1/shipping/manifests", handlers.carrierManifests)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/shipments", auth.PermissionRecordCreate, handlers.carrierManifestAddShipment)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/shipments/{shipment_id}", auth.PermissionRecordCreate, handlers.carrierManifestRemoveShipment)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/ready", auth.PermissionRecordCreate, handlers.carrierManifestReady)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/cancel", auth.PermissionRecordCreate, handlers.carrierManifestCancel)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/exceptions", auth.PermissionRecordCreate, handlers.carrierManifestExceptions)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/confirm-handover", auth.PermissionRecordCreate, handlers.carrierManifestHandover)
	routes.permission("/api/v1/shipping/manifests/{manifest_id}/scan", auth.PermissionShippingView, handlers.carrierManifestScan)
}
