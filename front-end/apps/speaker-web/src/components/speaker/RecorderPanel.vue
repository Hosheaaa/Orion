<template>
  <section class="recorder" aria-labelledby="recorder-heading">
    <header class="recorder__header">
      <div>
        <h2 id="recorder-heading">音频采集 · 设备状态</h2>
        <p>
          系统自动监测麦克风电平、采样率与实时抖动。点击“开始推流”后，音频将以 LINEAR16 16kHz 格式分帧发送至后端 WebSocket。
        </p>
      </div>
      <div class="recorder__status">
        <span class="status-dot" :class="statusClass" />
        <span>{{ statusText }}</span>
      </div>
    </header>
    <div class="recorder__body">
      <div class="recorder__visual">
        <div class="wave">
          <div
            v-for="index in 20"
            :key="index"
            class="wave__bar"
            :style="barStyle(index)"
          />
        </div>
        <div class="visual__meta">
          <div>
            <span class="visual__label">当前音量</span>
            <strong>{{ (store.micLevel * 100).toFixed(0) }}%</strong>
          </div>
          <div>
            <span class="visual__label">平均音量</span>
            <strong>63%</strong>
          </div>
          <div>
            <span class="visual__label">峰值</span>
            <strong>-3.2 dB</strong>
          </div>
        </div>
      </div>
      <aside class="recorder__controls">
        <div class="control-card">
          <header>
            <h3>采集参数</h3>
            <span>LINEAR16 · 16kHz · 单声道</span>
          </header>
          <ul>
            <li>
              <span>帧时长</span>
              <strong>160 ms</strong>
            </li>
            <li>
              <span>自动增益控制</span>
              <strong>已开启</strong>
            </li>
            <li>
              <span>噪声抑制</span>
              <strong>浏览器原生</strong>
            </li>
          </ul>
        </div>
        <div class="control-card">
          <header>
            <h3>设备检测</h3>
            <span>RØDE NT-USB Mini · 已接入</span>
          </header>
          <ul>
            <li>
              <span>延迟补偿</span>
              <strong>0.32 秒</strong>
            </li>
            <li>
              <span>缓冲深度</span>
              <strong>1.6 秒</strong>
            </li>
            <li>
              <span>最近抖动</span>
              <strong>1.2%</strong>
            </li>
          </ul>
        </div>
        <button
          class="recorder__cta"
          type="button"
          :class="{ 'is-live': store.isStreaming }"
          @click="toggleStreaming"
        >
          <span>{{ store.isStreaming ? "停止推流" : "开始推流" }}</span>
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              v-if="store.isStreaming"
              d="M8 7h3v10H8zm5 0h3v10h-3z"
              fill="currentColor"
            />
            <path
              v-else
              d="M8 6.5v11l9-5.5z"
              fill="currentColor"
            />
          </svg>
        </button>
        <button class="recorder__secondary" type="button">
          快速录制 30 秒彩排
        </button>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useSpeakerSessionStore } from "@/stores/speakerSession";

const store = useSpeakerSessionStore();

const statusText = computed(() =>
  store.isStreaming ? "推流中 · 正在向后端发送音频帧" : "待机 · 未开始推流"
);

const statusClass = computed(() =>
  store.isStreaming ? "is-live" : "is-idle"
);

function toggleStreaming() {
  if (store.isStreaming) {
    store.stopStreaming();
  } else {
    store.startStreaming();
  }
}

function barStyle(index: number) {
  const base = Math.sin((index / 20) * Math.PI);
  const scale = store.micLevel * 0.8 + 0.2;
  const height = 12 + Math.abs(base) * 48 * scale;
  const delay = index * 0.03;
  return {
    height: `${height}px`,
    animationDelay: `${delay}s`
  };
}
</script>

<style scoped>
.recorder {
  background: rgba(15, 23, 42, 0.92);
  color: white;
  border-radius: 26px;
  padding: 32px;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.recorder::before,
.recorder::after {
  content: "";
  position: absolute;
  inset: auto;
  width: 420px;
  height: 420px;
  filter: blur(140px);
  opacity: 0.55;
  pointer-events: none;
}

.recorder::before {
  top: -160px;
  left: -120px;
  background: rgba(16, 185, 129, 0.45);
}

.recorder::after {
  bottom: -180px;
  right: -140px;
  background: rgba(139, 92, 246, 0.32);
}

.recorder__header {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-start;
  position: relative;
  z-index: 1;
}

.recorder__header h2 {
  margin: 0 0 8px;
  font-size: 22px;
}

.recorder__header p {
  margin: 0;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(226, 232, 240, 0.82);
  max-width: 620px;
}

.recorder__status {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: rgba(15, 23, 42, 0.42);
  border: 1px solid rgba(148, 163, 184, 0.32);
  border-radius: 14px;
  padding: 10px 16px;
  font-size: 13px;
  font-weight: 600;
  color: rgba(226, 232, 240, 0.9);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(245, 245, 245, 0.4);
  box-shadow: 0 0 0 0 rgba(148, 163, 184, 0.6);
  transition: background 0.3s ease, box-shadow 0.3s ease;
}

.status-dot.is-live {
  background: #f43f5e;
  animation: pulse 1.4s infinite;
}

.status-dot.is-idle {
  background: rgba(148, 163, 184, 0.6);
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(244, 63, 94, 0.6);
  }
  70% {
    box-shadow: 0 0 0 16px rgba(244, 63, 94, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(244, 63, 94, 0);
  }
}

.recorder__body {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
  position: relative;
  z-index: 1;
}

.recorder__visual {
  background: rgba(15, 23, 42, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 22px;
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 26px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.wave {
  height: 140px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.wave__bar {
  width: 6px;
  border-radius: 12px;
  background: linear-gradient(180deg, rgba(56, 189, 248, 0.18), rgba(56, 189, 248, 0.65));
  transition: height 0.15s ease;
  animation: shimmer 1.4s ease-in-out infinite alternate;
}

@keyframes shimmer {
  to {
    transform: translateY(3px);
  }
}

.visual__meta {
  display: flex;
  gap: 32px;
}

.visual__meta > div {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.visual__label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: rgba(148, 163, 184, 0.8);
}

.visual__meta strong {
  font-size: 24px;
}

.recorder__controls {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.control-card {
  background: rgba(15, 23, 42, 0.52);
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 20px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.control-card header h3 {
  margin: 0;
  font-size: 16px;
}

.control-card header span {
  font-size: 12px;
  color: rgba(226, 232, 240, 0.65);
}

.control-card ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 10px;
}

.control-card li {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: rgba(226, 232, 240, 0.85);
}

.control-card strong {
  font-weight: 600;
  color: #38bdf8;
}

.recorder__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 16px;
  padding: 12px 18px;
  font-weight: 700;
  letter-spacing: 0.02em;
  border: 1px solid rgba(248, 113, 113, 0.42);
  background: rgba(248, 113, 113, 0.18);
  color: #fecaca;
  transition:
    transform 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.recorder__cta svg {
  width: 18px;
  height: 18px;
}

.recorder__cta:hover {
  transform: translateY(-3px);
  box-shadow:
    0 18px 38px -28px rgba(248, 113, 113, 0.75),
    0 1px 0 rgba(255, 255, 255, 0.6);
}

.recorder__cta.is-live {
  border-color: rgba(16, 185, 129, 0.45);
  background: rgba(16, 185, 129, 0.18);
  color: #bbf7d0;
}

.recorder__secondary {
  border-radius: 14px;
  padding: 12px 16px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(148, 163, 184, 0.18);
  color: rgba(226, 232, 240, 0.85);
  font-weight: 600;
  transition:
    transform 0.22s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.recorder__secondary:hover {
  transform: translateY(-2px);
  background: rgba(148, 163, 184, 0.24);
}

@media (max-width: 1024px) {
  .recorder__body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .recorder {
    padding: 24px;
  }

  .visual__meta {
    flex-direction: column;
    gap: 20px;
  }
}
</style>
