import { AxiosError } from "axios";
import type { AxiosResponse, InternalAxiosRequestConfig } from "axios";
import { vi } from "vitest";

import { api } from "@/lib/api";

/**
 * Test double for the *network*, not for the application. It replaces the
 * transport of the real `api` axios instance, so component tests still run
 * the real page, the real authService, the real request interceptors and the
 * real error handling - only "what the server answers" is scripted. Answers
 * mirror the backend's actual envelope: `{ success, message, code?, data? }`.
 */
export interface RecordedRequest {
  method: string;
  url: string;
  body: unknown;
}

export type MockReply = { status: number; body?: unknown } | { networkError: true };

export type MockHandler = (request: RecordedRequest) => MockReply;

export function installMockApi(handler: MockHandler): { requests: RecordedRequest[]; restore: () => void } {
  const requests: RecordedRequest[] = [];
  const originalAdapter = api.defaults.adapter;
  // The client logs failed requests; that is expected noise in these tests.
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  api.defaults.adapter = async (config: InternalAxiosRequestConfig) => {
    const request: RecordedRequest = {
      method: (config.method ?? "get").toUpperCase(),
      url: config.url ?? "",
      body: typeof config.data === "string" ? JSON.parse(config.data) : config.data,
    };
    requests.push(request);

    const reply = handler(request);
    if ("networkError" in reply) {
      throw new AxiosError("Network Error", AxiosError.ERR_NETWORK, config);
    }

    const response: AxiosResponse = {
      data: reply.body ?? {},
      status: reply.status,
      statusText: "",
      headers: {},
      config,
      request: {},
    };
    if (reply.status >= 200 && reply.status < 300) {
      return response;
    }
    throw new AxiosError(`Request failed with status code ${reply.status}`, AxiosError.ERR_BAD_REQUEST, config, {}, response);
  };

  return {
    requests,
    restore: () => {
      api.defaults.adapter = originalAdapter;
      consoleError.mockRestore();
    },
  };
}

/** The backend's error envelope. */
export function apiError(status: number, message: string, code?: string): MockReply {
  return { status, body: { success: false, message, ...(code ? { code } : {}) } };
}

export function apiOk(message: string, data?: unknown): MockReply {
  return { status: 200, body: { success: true, message, ...(data === undefined ? {} : { data }) } };
}
