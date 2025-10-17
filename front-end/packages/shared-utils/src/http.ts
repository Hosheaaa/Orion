/**
 * 简易 HTTP 客户端，封装 fetch 并自动附加鉴权信息。
 */
export interface HttpClientOptions {
  baseUrl: string;
  getAccessToken?: () => string | null;
  onUnauthorized?: () => void;
}

export interface RequestOptions extends RequestInit {
  query?: Record<string, string | number | boolean | undefined>;
}

export class HttpClient {
  private readonly baseUrl: string;
  private readonly getAccessToken?: () => string | null;
  private readonly onUnauthorized?: () => void;

  constructor(options: HttpClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.getAccessToken = options.getAccessToken;
    this.onUnauthorized = options.onUnauthorized;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const url = this.buildUrl(path, options.query);
    const headers = new Headers(options.headers);
    headers.set("Content-Type", "application/json");

    const token = this.getAccessToken?.();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (response.status === 401) {
      this.onUnauthorized?.();
      throw new Error("未授权访问，请重新登录。");
    }

    if (!response.ok) {
      const message = await this.safeParseError(response);
      throw new Error(message ?? `请求失败，状态码 ${response.status}`);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return (await response.json()) as T;
  }

  get<T>(path: string, options?: RequestOptions) {
    return this.request<T>(path, {
      ...options,
      method: "GET"
    });
  }

  post<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>(path, {
      ...options,
      method: "POST",
      body: body === undefined ? options?.body : JSON.stringify(body)
    });
  }

  put<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>(path, {
      ...options,
      method: "PUT",
      body: body === undefined ? options?.body : JSON.stringify(body)
    });
  }

  delete<T>(path: string, options?: RequestOptions) {
    return this.request<T>(path, {
      ...options,
      method: "DELETE"
    });
  }

  private buildUrl(path: string, query?: RequestOptions["query"]) {
    const url = new URL(path, this.baseUrl);
    if (query) {
      Object.entries(query).forEach(([key, value]) => {
        if (value === undefined) return;
        url.searchParams.set(key, String(value));
      });
    }
    return url.toString();
  }

  private async safeParseError(response: Response) {
    try {
      const data = await response.json();
      if (typeof data?.message === "string") {
        return data.message;
      }
      if (typeof data?.error === "string") {
        return data.error;
      }
      return null;
    } catch {
      return null;
    }
  }
}

export function createHttpClient(options: HttpClientOptions) {
  return new HttpClient(options);
}
