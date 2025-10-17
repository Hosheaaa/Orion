import { createHttpClient } from "@orion/shared-utils";
import { envConfig } from "@/config/env";
import { loadTokens, clearTokens } from "./tokenStorage";

let currentTokens = loadTokens();
const unauthorizedHandlers = new Set<() => void>();

export function setCurrentTokens(tokens: typeof currentTokens) {
  currentTokens = tokens;
}

export function onUnauthorized(handler: () => void) {
  unauthorizedHandlers.add(handler);
  return () => unauthorizedHandlers.delete(handler);
}

export const httpClient = createHttpClient({
  baseUrl: envConfig.apiBaseUrl,
  getAccessToken: () => currentTokens?.accessToken ?? null,
  onUnauthorized: () => {
    clearTokens();
    currentTokens = null;
    unauthorizedHandlers.forEach((handler) => handler());
  }
});
