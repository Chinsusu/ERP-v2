import type { WarehouseDailyBoardItem } from "../types";

export async function getWarehouseDailyBoard(): Promise<WarehouseDailyBoardItem[]> {
  return [
    { id: "receiving", label: "Receiving", count: 0, status: "normal" },
    { id: "qc-hold", label: "QC Hold", count: 0, status: "warning" },
    { id: "packing", label: "Packing", count: 0, status: "normal" },
    { id: "handover", label: "Carrier Handover", count: 0, status: "blocked" }
  ];
}
