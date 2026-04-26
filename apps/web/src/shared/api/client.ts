const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export type ApiErrorCode =
  | "VALIDATION_ERROR"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "NOT_FOUND"
  | "CONFLICT"
  | "INSUFFICIENT_STOCK"
  | "INVALID_STATE";

export type ApiSuccessResponse<T> = {
  success: true;
  data: T;
  request_id: string;
};

export type ApiErrorResponse = {
  error: {
    code: ApiErrorCode;
    message: string;
    details?: Record<string, unknown>;
    request_id: string;
  };
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: ApiErrorCode;
  readonly details?: Record<string, unknown>;
  readonly requestId: string;

  constructor(status: number, payload: ApiErrorResponse) {
    super(payload.error.message);
    this.name = "ApiError";
    this.status = status;
    this.code = payload.error.code;
    this.details = payload.error.details;
    this.requestId = payload.error.request_id;
  }
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`);
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<T>;
  return payload.data;
}

async function createApiError(response: Response): Promise<Error> {
  try {
    const payload = (await response.json()) as ApiErrorResponse;
    if (payload.error?.code && payload.error.message && payload.error.request_id) {
      return new ApiError(response.status, payload);
    }
  } catch {
    // Fall through to a generic error when the backend did not return the standard envelope.
  }

  return new Error(`API request failed: ${response.status}`);
}
