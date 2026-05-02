package main

import (
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func newRuntimeCarrierManifestStore(
	_ config.Config,
) (shippingapp.CarrierManifestStore, func() error, error) {
	return shippingapp.NewPrototypeCarrierManifestStore(), nil, nil
}

func newRuntimePickTaskStore(
	_ config.Config,
	tasks ...shippingdomain.PickTask,
) (shippingapp.PickTaskStore, func() error, error) {
	return shippingapp.NewPrototypePickTaskStore(tasks...), nil, nil
}

func newRuntimePackTaskStore(
	_ config.Config,
	tasks ...shippingdomain.PackTask,
) (shippingapp.PackTaskStore, func() error, error) {
	return shippingapp.NewPrototypePackTaskStore(tasks...), nil, nil
}
