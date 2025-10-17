<template>
  <section class="panel" aria-labelledby="activity-selector-heading">
    <header class="panel__header">
      <div>
        <h2 id="activity-selector-heading">活动排期 · 今日可选</h2>
        <p>
          系统自动匹配今日绑定的活动场次，选择后将同步加载目标语言、状态与观众端入口，并提醒对应的术语表。
        </p>
      </div>
      <button class="panel__ghost" type="button">查看历史场次</button>
    </header>
    <div class="panel__list">
      <article
        v-for="activity in activities"
        :key="activity.id"
        :class="['activity-card', { 'is-active': activity.id === selectedActivityId }]"
        @click="handleSelect(activity)"
        role="button"
        tabindex="0"
        @keyup.enter="handleSelect(activity)"
      >
        <header class="activity-card__header">
          <div>
            <h3>{{ activity.title }}</h3>
            <span class="activity-card__sub">
              {{ formatDate(activity.startTime) }} · {{ renderStatus(activity.status) }} · {{ activity.speaker }}
            </span>
          </div>
          <span class="activity-card__badge">
            输入语种：{{ formatLanguage(activity.inputLanguage) }}
          </span>
        </header>
        <p class="activity-card__description">
          {{ activity.description }}
        </p>
        <footer class="activity-card__footer">
          <div class="activity-card__langs">
            <span v-for="lang in activity.displayLanguages" :key="lang">{{ lang }}</span>
          </div>
          <div class="activity-card__cta">
            <span>进入彩排模式</span>
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path
                d="M7.5 4.5 12.5 10 7.5 15.5"
                fill="none"
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.8"
              />
            </svg>
          </div>
        </footer>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useSpeakerSessionStore } from "@/stores/speakerSession";
import type { ConsoleActivity } from "@/services/speakerConsoleService";

const props = defineProps<{ activities: ConsoleActivity[] }>();
const LANGUAGE_NAME_MAP: Record<string, string> = {
  "zh-CN": "简体中文",
  "zh-TW": "繁体中文",
  en: "英语",
  ja: "日语",
  ko: "韩语",
  es: "西班牙语",
  fr: "法语",
  de: "德语"
};

const store = useSpeakerSessionStore();

const selectedActivityId = computed(() => store.currentActivity?.id ?? "");

function handleSelect(activity: ConsoleActivity) {
  if (selectedActivityId.value === activity.id) return;
  store.selectActivity(activity);
}

function formatDate(iso: string) {
  const date = new Date(iso);
  return date.toLocaleString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit",
    day: "2-digit"
  });
}

function formatLanguage(code: string) {
  return LANGUAGE_NAME_MAP[code] ?? code;
}

function renderStatus(status: string) {
  switch (status) {
    case "draft":
      return "草稿";
    case "published":
      return "已发布";
    case "closed":
      return "已关闭";
    default:
      return status;
  }
}
</script>

<style scoped>
.panel {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  padding: 28px 30px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  box-shadow:
    0 22px 40px -30px rgba(15, 23, 42, 0.45),
    0 1px 0 rgba(255, 255, 255, 0.6);
}

.panel__header {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-start;
}

.panel__header h2 {
  margin: 0 0 8px;
  font-size: 22px;
  color: #0f172a;
}

.panel__header p {
  margin: 0;
  font-size: 14px;
  color: #475569;
  max-width: 600px;
}

.panel__ghost {
  border: 1px solid rgba(148, 163, 184, 0.38);
  background: rgba(248, 250, 252, 0.8);
  color: #1e293b;
  border-radius: 12px;
  padding: 10px 16px;
  font-weight: 600;
  transition:
    background 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.panel__ghost:hover {
  background: rgba(241, 245, 249, 0.95);
  transform: translateY(-2px);
  box-shadow: 0 12px 24px -18px rgba(15, 23, 42, 0.55);
}

.panel__list {
  display: grid;
  gap: 18px;
}

.activity-card {
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  padding: 24px 24px 22px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.92));
  cursor: pointer;
  transition:
    transform 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    border 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.activity-card:hover {
  transform: translateY(-4px);
  box-shadow:
    0 26px 38px -32px rgba(15, 23, 42, 0.55),
    0 18px 32px -30px rgba(20, 184, 166, 0.38);
  border-color: rgba(16, 185, 129, 0.65);
}

.activity-card.is-active {
  border-color: rgba(16, 185, 129, 0.9);
  box-shadow:
    0 28px 48px -34px rgba(20, 184, 166, 0.6),
    0 1px 0 rgba(255, 255, 255, 0.7);
  background: linear-gradient(145deg, rgba(224, 255, 244, 0.6), rgba(255, 255, 255, 0.95));
}

.activity-card__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.activity-card__header h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.activity-card__sub {
  display: inline-block;
  margin-top: 4px;
  font-size: 13px;
  color: #64748b;
}

.activity-card__badge {
  padding: 6px 10px;
  border-radius: 12px;
  background: rgba(14, 165, 233, 0.12);
  color: #0369a1;
  border: 1px solid rgba(14, 165, 233, 0.28);
  font-size: 12px;
  font-weight: 600;
}

.activity-card__description {
  margin: 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.65;
}

.activity-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.activity-card__langs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.activity-card__langs span {
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(30, 64, 175, 0.08);
  border: 1px solid rgba(30, 64, 175, 0.18);
  font-size: 12px;
  color: #1e3a8a;
}

.activity-card__cta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #0f172a;
  font-size: 13px;
  font-weight: 600;
  transition: color 0.24s ease;
}

.activity-card__cta svg {
  width: 16px;
  height: 16px;
}

.activity-card__cta:hover {
  color: #0ea371;
}

@media (max-width: 1024px) {
  .panel {
    padding: 24px;
  }

  .panel__header {
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .activity-card {
    padding: 20px;
  }

  .activity-card__footer {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>
