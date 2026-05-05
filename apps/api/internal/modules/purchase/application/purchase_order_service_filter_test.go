package application

import (
	"testing"

	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
)

func TestPurchaseOrderMatchesFilterSearchesNote(t *testing.T) {
	order := purchasedomain.PurchaseOrder{
		PONo:          "PO-260505-666046",
		SupplierCode:  "SUP-RM-BIO",
		SupplierName:  "BioActive Raw Materials",
		WarehouseCode: "WH-HCM-RM",
		Note:          "Tao tu ke hoach san xuat PP-260505-968033",
	}

	if !purchaseOrderMatchesFilter(order, PurchaseOrderFilter{Search: "PP-260505-968033"}) {
		t.Fatal("purchaseOrderMatchesFilter() = false, want search to match production plan note")
	}
}
