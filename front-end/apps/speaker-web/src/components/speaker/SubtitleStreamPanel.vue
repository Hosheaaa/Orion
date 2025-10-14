<template>
  <section class="subtitle-panel" aria-labelledby="subtitle-heading">
    <header class="subtitle-panel__header">
      <div>
        <h2 id="subtitle-heading">实时字幕流</h2>
        <p>
          最新字幕将自动固定在底部，系统会根据观众端订阅语言同步高亮。使用「观众视角预览」可校对译文呈现。
        </p>
      </div>
      <button class="subtitle-panel__action" type="button">
        观众视角预览
      </button>
    </header>
    <div class="subtitle-panel__stream">
      <div
        v-for="(subtitle, index) in items"
        :key="subtitle.id"
        :class="['subtitle-item', { 'is-latest': index === 0 }]"
      >
        <div class="subtitle-item__time">
          {{ formatTime(subtitle.timestamp) }}
        </div>
        <div class="subtitle-item__body">
          <div class="subtitle-item__original">
            {{ subtitle.original }}
          </div>
          <div class="subtitle-item__translated">
            {{ subtitle.translated }}
          </div>
        </div>
        <div class="subtitle-item__meta">
          <span class="tag tag--primary">同步推送</span>
          <span class="tag">观众端朗读可用</span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { SubtitleItem } from "@/stores/speakerSession";

defineProps<{
  items: SubtitleItem[];
}>();

function formatTime(iso: string) {
  const d = new Date(iso);
  return d.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  });
}
</script>

<style scoped>
.subtitle-panel {
  background: rgba(255, 255, 255, 0.94);
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 26px 30px;
  display: flex;
  flex-direction: column;
  gap: 22px;
  box-shadow:
    0 24px 44px -34px rgba(15, 23, 42, 0.5),
    0 1px 0 rgba(255, 255, 255, 0.65);
}

.subtitle-panel__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
}

.subtitle-panel__header h2 {
  margin: 0 0 6px;
  font-size: 20px;
  color: #0f172a;
}

.subtitle-panel__header p {
  margin: 0;
  font-size: 14px;
  color: #475569;
}

.subtitle-panel__action {
  padding: 10px 18px;
  border-radius: 14px;
  border: 1px solid rgba(139, 92, 246, 0.4);
  background: rgba(139, 92, 246, 0.12);
  color: #5b21b6;
  font-weight: 600;
  transition:
    transform 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.subtitle-panel__action:hover {
  transform: translateY(-2px);
  box-shadow: 0 20px 30px -28px rgba(88, 28, 135, 0.65);
}

.subtitle-panel__stream {
  display: grid;
  gap: 16px;
  max-height: 420px;
  overflow-y: auto;
  padding-right: 6px;
}

.subtitle-panel__stream::-webkit-scrollbar {
  width: 6px;
}

.subtitle-panel__stream::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.38);
  border-radius: 99px;
}

.subtitle-item {
  display: grid;
  grid-template-columns: 120px 1fr auto;
  gap: 18px;
  padding: 18px 22px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(248, 250, 252, 0.95), rgba(255, 255, 255, 0.98));
  border: 1px solid rgba(148, 163, 184, 0.22);
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.subtitle-item.is-latest {
  border-color: rgba(16, 185, 129, 0.7);
  box-shadow:
    0 24px 36px -30px rgba(16, 185, 129, 0.55),
    0 1px 0 rgba(255, 255, 255, 0.7);
  transform: translateY(-2px);
}

.subtitle-item__time {
  font-family: "JetBrains Mono", "SFMono-Regular", Consolas, monospace;
  font-size: 13px;
  color: #475569;
  letter-spacing: 0.05em;
  padding-top: 6px;
}

.subtitle-item__body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.subtitle-item__original {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  line-height: 1.6;
}

.subtitle-item__translated {
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
}

.subtitle-item__meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

.tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(148, 163, 184, 0.16);
}

.tag--primary {
  color: #047857;
  background: rgba(16, 185, 129, 0.16);
  border-color: rgba(16, 185, 129, 0.35);
}

@media (max-width: 1024px) {
  .subtitle-item {
    grid-template-columns: 1fr;
  }

  .subtitle-item__meta {
    flex-direction: row;
    justify-content: flex-start;
  }
}

@media (max-width: 768px) {
  .subtitle-panel {
    padding: 24px 20px;
  }

  .subtitle-panel__stream {
    max-height: 380px;
  }
}
</style>
