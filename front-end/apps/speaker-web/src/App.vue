<template>
  <n-config-provider
    :theme-overrides="themeOverrides"
    :theme="lightTheme"
    abstract
  >
    <div v-if="showShell" class="app-shell">
      <header class="app-shell__header">
        <div class="brand">
          <div class="brand__mark">Orion Live</div>
          <span class="brand__tagline">{{ activeNavLabel }}</span>
        </div>
        <nav class="primary-nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.name"
            :to="{ name: item.name }"
            class="primary-nav__item"
            :class="{ 'is-active': route.name === item.name }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
        <div class="user-chip">
          <span class="user-chip__avatar">{{ userInitial }}</span>
          <div class="user-chip__meta">
            <span class="user-chip__name">{{ userName }}</span>
            <span class="user-chip__session">{{ activeNavLabel }}</span>
          </div>
        </div>
      </header>
      <main class="app-shell__main">
        <router-view />
      </main>
    </div>
    <router-view v-else />
  </n-config-provider>
</template>

<script setup lang="ts">
import { lightTheme } from "naive-ui";
import type { GlobalThemeOverrides } from "naive-ui";
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { storeToRefs } from "pinia";

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: "#10b981",
    primaryColorHover: "#0ea371",
    primaryColorPressed: "#0b8d63",
    primaryColorSuppl: "#34d399"
  }
};

const route = useRoute();
const auth = useAuthStore();
const { profile, isAuthenticated } = storeToRefs(auth);

const navItems = [
  {
    name: "speaker-console",
    label: "演讲者面板"
  }
];

const showShell = computed(
  () => isAuthenticated.value && route.meta.requiresAuth !== false && route.meta.shell !== false
);
const userName = computed(() => profile.value?.username ?? "未登录");
const userInitial = computed(() => userName.value.slice(0, 1).toUpperCase());
const activeNavLabel = computed(() => {
  const current = navItems.find((item) => item.name === route.name);
  return current?.label ?? "实时多语演讲控制台";
});
</script>
