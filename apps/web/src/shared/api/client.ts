import type { components, paths } from "./generated/schema";
import { resolveApiAccessToken } from "../auth/clientSessionToken";

const defaultApiBaseUrl = "/api/v1";

export type ApiErrorCode = components["schemas"]["ErrorCode"];

type GetOperation<Path extends keyof paths> = paths[Path] extends { get: infer Operation } ? Operation : never;

export type ApiGetPath = {
  [Path in keyof paths]: GetOperation<Path> extends never ? never : Path;
}[keyof paths];

type JsonSuccessResponse<Operation> = Operation extends {
  responses: { 200: { content: { "application/json": infer Response } } };
}
  ? Response
  : never;

type ApiSuccessData<Response> = Response extends { data: infer Data } ? Data : never;

type QueryParameters<Operation> = Operation extends { parameters: { query?: infer Query } } ? Query : never;

export type ApiGetResponse<Path extends ApiGetPath> = ApiSuccessData<JsonSuccessResponse<GetOperation<Path>>>;

export type ApiGetQuery<Path extends ApiGetPath> = [QueryParameters<GetOperation<Path>>] extends [never]
  ? undefined
  : QueryParameters<GetOperation<Path>>;

export type ApiSuccessResponse<T> = {
  success: true;
  data: T;
  request_id: string;
};

export type ApiErrorResponse = components["schemas"]["ErrorResponse"];

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

export type ApiRequestOptions<Path extends ApiGetPath> = {
  accessToken?: string;
  query?: ApiGetQuery<Path>;
};

export type ApiWriteOptions = {
  accessToken?: string;
};

export async function apiGet<Path extends ApiGetPath>(
  path: Path,
  options: ApiRequestOptions<Path> = {}
): Promise<ApiGetResponse<Path>> {
  const response = await fetch(apiUrl(path, options.query), {
    headers: authHeaders(options)
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<ApiGetResponse<Path>>;
  return payload.data;
}

export async function apiGetRaw<TData>(path: string, options: ApiWriteOptions = {}): Promise<TData> {
  const response = await fetch(apiUrl(path), {
    headers: authHeaders(options)
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<TData>;
  return payload.data;
}

export async function apiGetBlob(
  path: string,
  options: ApiWriteOptions = {}
): Promise<{ blob: Blob; filename?: string }> {
  const response = await fetch(apiUrl(path), {
    headers: authHeaders(options)
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  return {
    blob: await response.blob(),
    filename: filenameFromContentDisposition(response.headers.get("Content-Disposition"))
  };
}

export async function apiPost<TData, TBody>(
  path: string,
  body: TBody,
  options: ApiWriteOptions = {}
): Promise<TData> {
  return apiWrite<TData, TBody>("POST", path, body, options);
}

export async function apiPostForm<TData>(
  path: string,
  body: FormData,
  options: ApiWriteOptions = {}
): Promise<TData> {
  const response = await fetch(apiUrl(path), {
    method: "POST",
    headers: authHeaders(options),
    body
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<TData>;
  return payload.data;
}

export async function apiPatch<TData, TBody>(
  path: string,
  body: TBody,
  options: ApiWriteOptions = {}
): Promise<TData> {
  return apiWrite<TData, TBody>("PATCH", path, body, options);
}

export async function apiDelete<TData>(path: string, options: ApiWriteOptions = {}): Promise<TData> {
  const response = await fetch(apiUrl(path), {
    method: "DELETE",
    headers: authHeaders(options)
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<TData>;
  return payload.data;
}

function authHeaders(options: { accessToken?: string }) {
  const accessToken = resolveApiAccessToken(options.accessToken);
  if (!accessToken) {
    return undefined;
  }

  return {
    Authorization: `Bearer ${accessToken}`
  };
}

async function apiWrite<TData, TBody>(
  method: "PATCH" | "POST",
  path: string,
  body: TBody,
  options: ApiWriteOptions
): Promise<TData> {
  const response = await fetch(apiUrl(path), {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(authHeaders(options) ?? {})
    },
    body: JSON.stringify(body)
  });
  if (!response.ok) {
    throw await createApiError(response);
  }

  const payload = (await response.json()) as ApiSuccessResponse<TData>;
  return payload.data;
}

function apiUrl(path: string, query?: unknown) {
  return `${resolveApiBaseUrl()}${path}${queryString(query)}`;
}

function resolveApiBaseUrl() {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || defaultApiBaseUrl;
  const baseUrl = trimTrailingSlash(configuredBaseUrl);

  if (currentLocationProtocol() === "https:" && baseUrl.toLowerCase().startsWith("http://")) {
    try {
      return normalizeApiBasePath(new URL(baseUrl).pathname);
    } catch {
      return defaultApiBaseUrl;
    }
  }

  return baseUrl;
}

function normalizeApiBasePath(pathname: string) {
  const path = trimTrailingSlash(pathname.trim() || defaultApiBaseUrl);
  return path.startsWith("/") ? path : `/${path}`;
}

function trimTrailingSlash(value: string) {
  return value.length > 1 && value.endsWith("/") ? value.slice(0, -1) : value;
}

function currentLocationProtocol() {
  if (typeof window !== "undefined") {
    return window.location.protocol;
  }
  if (typeof globalThis.location !== "undefined") {
    return globalThis.location.protocol;
  }

  return "";
}

function queryString(query: unknown) {
  if (!query || typeof query !== "object") {
    return "";
  }

  const params = new URLSearchParams();
  Object.entries(query as Record<string, boolean | number | string | null | undefined>).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });

  const value = params.toString();
  return value ? `?${value}` : "";
}

function filenameFromContentDisposition(value: string | null) {
  if (!value) {
    return undefined;
  }

  const match = /filename="([^"]+)"/.exec(value);
  return match?.[1];
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
