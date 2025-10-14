<template>
  <n-config-provider
    :theme-overrides="themeOverrides"
    :theme="lightTheme"
    abstract
  >
    <div class="app-shell">
      <header class="app-shell__header">
        <div class="brand">
          <div class="brand__mark">Orion Live</div>
          <span class="brand__tagline">实时多语演讲控制台</span>
        </div>
        <nav class="primary-nav">
          <a class="primary-nav__item is-active" href="#">演讲者面板</a>
          <a class="primary-nav__item" href="#">观众概览</a>
          <a class="primary-nav__item" href="#">设置</a>
        </nav>
        <div class="user-chip">
          <span class="user-chip__avatar">陈</span>
          <div class="user-chip__meta">
            <span class="user-chip__name">陈述 • 主讲人</span>
            <span class="user-chip__session">GMT+8 · 创新发布会</span>
          </div>
        </div>
      </header>
      <main class="app-shell__main">
        <router-view />
      </main>
    </div>
  </n-config-provider>
</template>

<script setup lang="ts">
import { lightTheme } from "naive-ui";
import type { GlobalThemeOverrides } from "naive-ui";

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: "#10b981",
    primaryColorHover: "#0ea371",
    primaryColorPressed: "#0b8d63",
    primaryColorSuppl: "#34d399"
  }
};
</script>
