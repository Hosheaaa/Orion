<template>
  <div class="console">
    <hero-metrics-strip v-if="heroMetrics.length" :metrics="heroMetrics" />
    <section class="console__grid">
      <div class="console__primary">
        <activity-selection-panel
          v-if="activities.length"
          :activities="activities"
        />
        <recorder-panel />
        <subtitle-stream-panel
          :items="subtitleFeed"
        />
      </div>
      <aside class="console__secondary">
        <connection-status-panel
          v-if="connection"
          :snapshot="connection"
        />
        <preparation-checklist
          v-if="guidance.length"
          :items="guidance"
        />
        <section class="support-card">
          <header>
            <h2>快速帮助</h2>
            <p>
              直播期间如遇异常，可一键通知后台技术值班。我们会自动附带最近 2 分钟的诊断数据。
            </p>
          </header>
          <ul>
            <li>
              <span>WebSocket 状态广播</span>
              <strong>正常</strong>
            </li>
            <li>
              <span>翻译服务可用性</span>
              <strong>99.98%</strong>
            </li>
            <li>
              <span>当前观众端活跃语言</span>
              <strong>英语 / 日语 / 西语</strong>
            </li>
          </ul>
          <div class="support-card__cta">
            <button type="button">通知技术值班</button>
            <button type="button" class="ghost">查看应急预案</button>
          </div>
        </section>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from "vue";
import { useIntervalFn } from "@vueuse/core";
import HeroMetricsStrip from "@/components/speaker/HeroMetricsStrip.vue";
import ActivitySelectionPanel from "@/components/speaker/ActivitySelectionPanel.vue";
import RecorderPanel from "@/components/speaker/RecorderPanel.vue";
import SubtitleStreamPanel from "@/components/speaker/SubtitleStreamPanel.vue";
import ConnectionStatusPanel from "@/components/speaker/ConnectionStatusPanel.vue";
import PreparationChecklist from "@/components/speaker/PreparationChecklist.vue";
import {
  useHeroInsights,
  useSpeakerActivities,
  useSubtitleHistory,
  useGuidanceChecklist
} from "@/composables/useSpeakerConsoleData";
import { useSpeakerSessionStore } from "@/stores/speakerSession";
import type { SubtitleItem } from "@/stores/speakerSession";
import type { ConsoleActivity } from "@/services/speakerConsoleService";

const store = useSpeakerSessionStore();

const { data: activitiesData } = useSpeakerActivities();
const { data: heroMetricsData } = useHeroInsights();
const { data: subtitleHistoryData } = useSubtitleHistory();
const { data: guidanceData } = useGuidanceChecklist();

watch(activitiesData, (payload) => {
  if (!payload?.length) return;
  if (!store.currentActivity) {
    store.selectActivity(payload[0]);
  }
});

watch(subtitleHistoryData, (payload) => {
  if (!payload?.length) return;
  payload.slice(0, 3).forEach((item) => store.pushSubtitle(item));
});

const { pause, resume } = useIntervalFn(() => {
  const variance = Math.random() * 0.35;
  store.updateMicLevel(Math.min(1, 0.3 + variance));
}, 380, { immediate: true });

watch(
  () => store.isStreaming,
  (isStreaming) => {
    if (isStreaming) {
      pause();
    } else {
      store.updateMicLevel(0.2);
      resume();
    }
  },
  { immediate: true }
);

const activities = computed<ConsoleActivity[]>(() => activitiesData.value ?? []);
const heroMetrics = computed(() => heroMetricsData.value ?? []);
const connection = computed(() => store.connection);
const guidance = computed(() => guidanceData.value ?? []);

const subtitleFeed = computed<SubtitleItem[]>(() => {
  if (store.subtitles.length) {
    return store.subtitles;
  }
  return subtitleHistoryData.value ?? [];
});
</script>

<style scoped>
.console {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.console__grid {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 3fr);
  gap: 24px;
  align-items: start;
}

.console__primary {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.console__secondary {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.support-card {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 24px 26px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  box-shadow:
    0 20px 40px -34px rgba(15, 23, 42, 0.45),
    0 1px 0 rgba(255, 255, 255, 0.65);
}

.support-card header h2 {
  margin: 0 0 6px;
  font-size: 18px;
  color: #0f172a;
}

.support-card header p {
  margin: 0;
  font-size: 14px;
  color: #475569;
}

.support-card ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}

.support-card li {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #475569;
  padding: 10px 12px;
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.support-card strong {
  color: #0f172a;
  font-weight: 600;
}

.support-card__cta {
  display: flex;
  gap: 12px;
  justify-content: flex-start;
  flex-wrap: wrap;
}

.support-card__cta button {
  padding: 10px 16px;
  border-radius: 12px;
  border: 1px solid rgba(16, 185, 129, 0.4);
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
  font-weight: 600;
  transition:
    transform 0.22s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.support-card__cta button:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 24px -22px rgba(16, 185, 129, 0.6);
}

.support-card__cta .ghost {
  border-color: rgba(148, 163, 184, 0.4);
  background: rgba(148, 163, 184, 0.16);
  color: #1e293b;
}

@media (max-width: 1200px) {
  .console__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .console {
    gap: 24px;
  }
}
</style>
