import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { login as loginRequest, refresh as refreshRequest } from "@/services/authService";
import { clearTokens, loadTokens, saveTokens, type StoredTokens } from "@/services/tokenStorage";
import { setCurrentTokens, onUnauthorized } from "@/services/httpClient";

interface LoginForm {
  username: string;
  password: string;
}

function calcExpiry(expiresIn: number) {
  return Date.now() + expiresIn * 1000;
}

export const useAuthStore = defineStore("auth", () => {
  const tokens = ref<StoredTokens | null>(loadTokens());
  const profile = ref<{ username: string } | null>(
    tokens.value?.username ? { username: tokens.value.username } : null
  );
  const loading = ref(false);

  if (tokens.value) {
    setCurrentTokens(tokens.value);
  }

  onUnauthorized(() => {
    tokens.value = null;
    profile.value = null;
  });

  const isAuthenticated = computed(() => !!tokens.value);

  async function login(form: LoginForm) {
    loading.value = true;
    try {
      const response = await loginRequest(form);
      const stored: StoredTokens = {
        accessToken: response.accessToken,
        refreshToken: response.refreshToken,
        expiresAt: calcExpiry(response.expiresIn),
        username: form.username
      };
      tokens.value = stored;
      profile.value = {
        username: form.username
      };
      saveTokens(stored);
      setCurrentTokens(stored);
    } finally {
      loading.value = false;
    }
  }

  async function refresh() {
    if (!tokens.value) return;
    const currentUsername = tokens.value.username;
    const response = await refreshRequest(tokens.value.refreshToken);
    const stored: StoredTokens = {
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
      expiresAt: calcExpiry(response.expiresIn),
      username: currentUsername
    };
    tokens.value = stored;
    saveTokens(stored);
    setCurrentTokens(stored);
    if (currentUsername) {
      profile.value = { username: currentUsername };
    }
  }

  function logout() {
    tokens.value = null;
    profile.value = null;
    clearTokens();
    setCurrentTokens(null);
  }

  function shouldRefresh() {
    if (!tokens.value) return false;
    const remain = tokens.value.expiresAt - Date.now();
    return remain < 60_000; // 小于 1 分钟则刷新
  }

  return {
    tokens,
    profile,
    loading,
    isAuthenticated,
    login,
    refresh,
    shouldRefresh,
    logout
  };
});
