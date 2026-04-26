export type WarehouseDailyBoardItem = {
  id: string;
  label: string;
  count: number;
  status: "normal" | "warning" | "blocked";
};
