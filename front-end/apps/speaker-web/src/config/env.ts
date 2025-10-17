/**
 * 统一读取环境变量，便于集中校验。
 */
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
const wsBaseUrl = import.meta.env.VITE_WS_BASE_URL;

if (!apiBaseUrl) {
  console.warn("[Orion] 未配置 VITE_API_BASE_URL，默认回退为 http://localhost:8080");
}

export const envConfig = {
  apiBaseUrl: apiBaseUrl ?? "http://localhost:8080",
  wsBaseUrl: resolveWsBaseUrl(wsBaseUrl ?? apiBaseUrl ?? "http://localhost:8080")
};

function resolveWsBaseUrl(base: string) {
  if (!base) {
    return "ws://localhost:8080";
  }
  if (base.startsWith("ws://") || base.startsWith("wss://")) {
    return base.replace(/\/$/, "");
  }
  if (base.startsWith("https://")) {
    return base.replace("https://", "wss://").replace(/\/$/, "");
  }
  if (base.startsWith("http://")) {
    return base.replace("http://", "ws://").replace(/\/$/, "");
  }
  return base.replace(/\/$/, "");
}
