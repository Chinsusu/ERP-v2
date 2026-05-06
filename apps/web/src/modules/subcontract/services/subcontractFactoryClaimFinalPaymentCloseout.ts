import type {
  SubcontractFactoryClaim,
  SubcontractOrder,
  SubcontractOrderStatus
} from "../types";

export type FactoryClaimCloseoutStatus = "complete" | "current" | "pending" | "blocked";

export type SubcontractFactoryClaimFinalPaymentCloseout = {
  latestClaim?: SubcontractFactoryClaim;
  blockingClaimCount: number;
  claimStatus: FactoryClaimCloseoutStatus;
  finalPaymentStatus: FactoryClaimCloseoutStatus;
  canAcknowledgeClaim: boolean;
  canResolveClaim: boolean;
  canMarkFinalPaymentReady: boolean;
  finalPaymentBlockedReason?: string;
};

export function buildSubcontractFactoryClaimFinalPaymentCloseout(
  order: SubcontractOrder,
  claims: SubcontractFactoryClaim[] = []
): SubcontractFactoryClaimFinalPaymentCloseout {
  const sortedClaims = [...claims].sort((left, right) => right.openedAt.localeCompare(left.openedAt));
  const latestClaim = sortedClaims[0];
  const blockingClaimCount = sortedClaims.filter((claim) => claim.blocksFinalPayment && isBlockingClaimStatus(claim.status)).length;
  const fullFactoryReject = order.status === "rejected_with_factory_issue" && numeric(order.acceptedQty) <= 0;
  const accepted = order.status === "accepted";
  const finalPaymentReady = ["final_payment_ready", "closed"].includes(order.status);
  const claimStatus = claimStepStatus(order.status, latestClaim, blockingClaimCount);
  const finalPaymentStatus = finalPaymentStepStatus(order.status, blockingClaimCount, fullFactoryReject);

  return {
    latestClaim,
    blockingClaimCount,
    claimStatus,
    finalPaymentStatus,
    canAcknowledgeClaim: latestClaim?.status === "open",
    canResolveClaim: latestClaim ? ["open", "acknowledged"].includes(latestClaim.status) : false,
    canMarkFinalPaymentReady: accepted && blockingClaimCount === 0 && !fullFactoryReject,
    finalPaymentBlockedReason: finalPaymentBlockedReason(order, blockingClaimCount, fullFactoryReject, finalPaymentReady)
  };
}

function claimStepStatus(
  orderStatus: SubcontractOrderStatus,
  latestClaim: SubcontractFactoryClaim | undefined,
  blockingClaimCount: number
): FactoryClaimCloseoutStatus {
  if (!latestClaim) {
    return ["accepted", "final_payment_ready", "closed"].includes(orderStatus) ? "complete" : "pending";
  }
  if (blockingClaimCount > 0) {
    return "current";
  }

  return ["resolved", "closed", "cancelled"].includes(latestClaim.status) ? "complete" : "current";
}

function finalPaymentStepStatus(
  orderStatus: SubcontractOrderStatus,
  blockingClaimCount: number,
  fullFactoryReject: boolean
): FactoryClaimCloseoutStatus {
  if (["final_payment_ready", "closed"].includes(orderStatus)) {
    return "complete";
  }
  if (fullFactoryReject || blockingClaimCount > 0) {
    return "blocked";
  }
  if (orderStatus === "accepted") {
    return "current";
  }

  return "pending";
}

function finalPaymentBlockedReason(
  order: SubcontractOrder,
  blockingClaimCount: number,
  fullFactoryReject: boolean,
  finalPaymentReady: boolean
) {
  if (finalPaymentReady) {
    return undefined;
  }
  if (fullFactoryReject) {
    return "QC lỗi toàn bộ; cần xử lý đổi/bù hàng hoặc tạo lệnh mới trước khi thanh toán cuối.";
  }
  if (blockingClaimCount > 0) {
    return `${blockingClaimCount} claim nhà máy còn mở; phải xác nhận/chốt xử lý trước khi mở thanh toán cuối.`;
  }
  if (order.status !== "accepted") {
    return "Chỉ mở thanh toán cuối sau khi thành phẩm đã được QC chấp nhận.";
  }

  return undefined;
}

function isBlockingClaimStatus(status: SubcontractFactoryClaim["status"]) {
  return status === "open" || status === "acknowledged";
}

function numeric(value?: string) {
  if (!value) {
    return 0;
  }

  return Number.parseFloat(value.replace(",", ".")) || 0;
}
