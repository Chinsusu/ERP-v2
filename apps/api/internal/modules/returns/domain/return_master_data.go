package domain

import "sort"

type ReturnReason struct {
	Code        string
	Label       string
	Description string
	Active      bool
	SortOrder   int
}

type ReturnCondition struct {
	Code                 string
	Label                string
	Description          string
	DefaultDisposition   ReturnDisposition
	InventoryDisposition string
	RequiresQA           bool
	Active               bool
	SortOrder            int
}

type ReturnDispositionMaster struct {
	Code                  ReturnDisposition
	Label                 string
	Description           string
	InventoryDisposition  string
	TargetStockStatus     string
	TargetLocationType    string
	CreatesAvailableStock bool
	RequiresApproval      bool
	Active                bool
	SortOrder             int
}

type ReturnMasterData struct {
	Reasons      []ReturnReason
	Conditions   []ReturnCondition
	Dispositions []ReturnDispositionMaster
}

func PrototypeReturnMasterData() ReturnMasterData {
	data := ReturnMasterData{
		Reasons: []ReturnReason{
			{
				Code:        "failed_delivery",
				Label:       "Failed delivery",
				Description: "Carrier or shipper could not complete delivery.",
				Active:      true,
				SortOrder:   10,
			},
			{
				Code:        "customer_refused",
				Label:       "Customer refused",
				Description: "Customer refused the order at handover.",
				Active:      true,
				SortOrder:   20,
			},
			{
				Code:        "wrong_item",
				Label:       "Wrong item",
				Description: "Returned item does not match the original order line.",
				Active:      true,
				SortOrder:   30,
			},
			{
				Code:        "damaged_package",
				Label:       "Damaged package",
				Description: "Outer package or product arrived damaged.",
				Active:      true,
				SortOrder:   40,
			},
			{
				Code:        "quality_issue",
				Label:       "Quality issue",
				Description: "Customer or warehouse reported a suspected quality issue.",
				Active:      true,
				SortOrder:   50,
			},
			{
				Code:        "other",
				Label:       "Other",
				Description: "Manual review is required before disposition.",
				Active:      true,
				SortOrder:   90,
			},
		},
		Conditions: []ReturnCondition{
			{
				Code:                 "sealed_good",
				Label:                "Sealed good",
				Description:          "Seal is intact and item appears sellable after inspection.",
				DefaultDisposition:   ReturnDispositionReusable,
				InventoryDisposition: "restock_available",
				RequiresQA:           false,
				Active:               true,
				SortOrder:            10,
			},
			{
				Code:                 "opened_good",
				Label:                "Opened good",
				Description:          "Package was opened but item appears usable.",
				DefaultDisposition:   ReturnDispositionNeedsInspection,
				InventoryDisposition: "hold_investigation",
				RequiresQA:           false,
				Active:               true,
				SortOrder:            20,
			},
			{
				Code:                 "damaged",
				Label:                "Damaged",
				Description:          "Item or packaging is damaged and cannot return to available stock.",
				DefaultDisposition:   ReturnDispositionNotReusable,
				InventoryDisposition: "scrap",
				RequiresQA:           true,
				Active:               true,
				SortOrder:            30,
			},
			{
				Code:                 "expired",
				Label:                "Expired",
				Description:          "Expiry date has passed or item is not sellable by shelf-life rule.",
				DefaultDisposition:   ReturnDispositionNotReusable,
				InventoryDisposition: "scrap",
				RequiresQA:           true,
				Active:               true,
				SortOrder:            40,
			},
			{
				Code:                 "suspected_quality_issue",
				Label:                "Suspected quality issue",
				Description:          "Item needs QA or lab review before any stock decision.",
				DefaultDisposition:   ReturnDispositionNeedsInspection,
				InventoryDisposition: "restock_quarantine",
				RequiresQA:           true,
				Active:               true,
				SortOrder:            50,
			},
		},
		Dispositions: []ReturnDispositionMaster{
			{
				Code:                  ReturnDispositionReusable,
				Label:                 "Reusable",
				Description:           "Potentially reusable after inspection; receiving keeps stock pending.",
				InventoryDisposition:  "restock_available",
				TargetStockStatus:     "return_pending",
				TargetLocationType:    "return_receiving",
				CreatesAvailableStock: false,
				RequiresApproval:      false,
				Active:                true,
				SortOrder:             10,
			},
			{
				Code:                  ReturnDispositionNotReusable,
				Label:                 "Not reusable",
				Description:           "Cannot return to sellable stock; route to lab, defect, or scrap.",
				InventoryDisposition:  "scrap",
				TargetStockStatus:     "damaged",
				TargetLocationType:    "defect_or_lab",
				CreatesAvailableStock: false,
				RequiresApproval:      false,
				Active:                true,
				SortOrder:             20,
			},
			{
				Code:                  ReturnDispositionNeedsInspection,
				Label:                 "Needs inspection",
				Description:           "Requires inspection or QA hold before stock movement.",
				InventoryDisposition:  "restock_quarantine",
				TargetStockStatus:     "quarantine",
				TargetLocationType:    "return_inspection",
				CreatesAvailableStock: false,
				RequiresApproval:      true,
				Active:                true,
				SortOrder:             30,
			},
		},
	}

	SortReturnMasterData(&data)

	return data
}

func SortReturnMasterData(data *ReturnMasterData) {
	sort.SliceStable(data.Reasons, func(i int, j int) bool {
		return data.Reasons[i].SortOrder < data.Reasons[j].SortOrder
	})
	sort.SliceStable(data.Conditions, func(i int, j int) bool {
		return data.Conditions[i].SortOrder < data.Conditions[j].SortOrder
	})
	sort.SliceStable(data.Dispositions, func(i int, j int) bool {
		return data.Dispositions[i].SortOrder < data.Dispositions[j].SortOrder
	})
}
