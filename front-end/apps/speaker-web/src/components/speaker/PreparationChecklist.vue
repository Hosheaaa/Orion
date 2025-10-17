<template>
  <section class="checklist" aria-labelledby="checklist-heading">
    <header>
      <h2 id="checklist-heading">演讲提醒 · 保持节奏</h2>
      <p>
        根据历次演讲的语速与字幕表现生成优化建议，彩排时逐项确认，可显著提升实时翻译准确率与观众体验。
      </p>
    </header>
    <ul>
      <li
        v-for="item in items"
        :key="item.title"
        :class="['checklist__item', `is-${item.emphasis}`]"
      >
        <div class="checklist__icon" aria-hidden="true">
          <svg viewBox="0 0 24 24">
            <path
              d="M6 12.5 10 16l8-8"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
              fill="none"
            />
          </svg>
        </div>
        <div>
          <h3>{{ item.title }}</h3>
          <p>{{ item.detail }}</p>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import type { GuidanceChecklistItem } from "@/services/speakerConsoleService";

defineProps<{
  items: GuidanceChecklistItem[];
}>();
</script>

<style scoped>
.checklist {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow:
    0 20px 40px -34px rgba(15, 23, 42, 0.45),
    0 1px 0 rgba(255, 255, 255, 0.7);
}

.checklist header h2 {
  margin: 0 0 6px;
  font-size: 18px;
  color: #0f172a;
}

.checklist header p {
  margin: 0;
  color: #475569;
  font-size: 14px;
}

.checklist ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 16px;
}

.checklist__item {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 16px;
  border-radius: 18px;
  padding: 18px 20px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: linear-gradient(135deg, rgba(248, 250, 252, 0.95), rgba(255, 255, 255, 0.98));
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.checklist__item:hover {
  transform: translateY(-2px);
  box-shadow:
    0 18px 30px -28px rgba(15, 23, 42, 0.55),
    0 12px 24px -18px rgba(20, 184, 166, 0.35);
}

.checklist__icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: rgba(16, 185, 129, 0.12);
  color: #0f766e;
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.checklist__item.is-success .checklist__icon {
  background: rgba(14, 165, 233, 0.12);
  color: #0369a1;
  border-color: rgba(14, 165, 233, 0.28);
}

.checklist__item.is-warning .checklist__icon {
  background: rgba(249, 115, 22, 0.15);
  color: #b45309;
  border-color: rgba(249, 115, 22, 0.32);
}

.checklist__item h3 {
  margin: 2px 0 6px;
  font-size: 16px;
  color: #0f172a;
}

.checklist__item p {
  margin: 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .checklist {
    padding: 20px;
  }
}
</style>
