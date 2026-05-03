import { NextResponse } from "next/server";
import { refreshCurrentBackendSession } from "@/shared/auth/serverSession";

export async function POST() {
  const result = await refreshCurrentBackendSession();
  if (!result.ok) {
    return NextResponse.json({ success: false }, { status: 401 });
  }

  return NextResponse.json({
    success: true,
    data: {
      access_token: result.accessToken,
      expires_at: result.expiresAt
    }
  });
}
