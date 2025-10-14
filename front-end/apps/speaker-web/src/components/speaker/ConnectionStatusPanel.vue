<template>
  <section class="connection-panel" aria-labelledby="connection-heading">
    <header class="connection-panel__header">
      <div>
        <h2 id="connection-heading">连接健康度</h2>
        <p>
          系统每 20 秒刷新一次心跳详情。当延迟或丢包超过阈值时，将自动提示切换网络或重新鉴权，确保长场次推流稳定。
        </p>
      </div>
      <span class="connection-panel__pill" :class="statusClass">
        {{ statusText }}
      </span>
    </header>
    <dl class="connection-panel__grid">
      <div>
        <dt>WebSocket 地址</dt>
        <dd>{{ snapshot.websocketUrl }}</dd>
      </div>
      <div>
        <dt>往返延迟</dt>
        <dd>{{ snapshot.latencyMs }} ms</dd>
      </div>
      <div>
        <dt>丢包率</dt>
        <dd>{{ snapshot.packetLossRate }}%</dd>
      </div>
      <div>
        <dt>重连次数</dt>
        <dd>{{ snapshot.reconnectAttempts }} 次</dd>
      </div>
      <div>
        <dt>最近心跳</dt>
        <dd>{{ formatTime(snapshot.lastHeartbeatAt) }}</dd>
      </div>
    </dl>
    <footer class="connection-panel__footer">
      <div>
        <strong>稳定性建议</strong>
        <p>
          当前网络稳定，建议保持有线连接；若延迟 > 180ms，将自动降级帧长至 160ms 并提示暂停切换网络。
        </p>
      </div>
      <button type="button">导出诊断日志</button>
    </footer>
  </section>
</template>

<script setup lang="ts">
import type { ConnectionSnapshot } from "@/stores/speakerSession";
import { computed } from "vue";

const props = defineProps<{
  snapshot: ConnectionSnapshot;
}>();

const statusClass = computed(() => `is-${props.snapshot.status}`);

const statusText = computed(() => {
  switch (props.snapshot.status) {
    case "connected":
      return "连接稳定";
    case "reconnecting":
      return "正在重连";
    case "degraded":
      return "网络波动";
    default:
      return "连接未知";
  }
});

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
.connection-panel {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  padding: 26px 30px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  box-shadow:
    0 22px 42px -34px rgba(15, 23, 42, 0.5),
    0 1px 0 rgba(255, 255, 255, 0.7);
}

.connection-panel__header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-start;
}

.connection-panel__header h2 {
  margin: 0 0 6px;
  font-size: 20px;
  color: #0f172a;
}

.connection-panel__header p {
  margin: 0;
  color: #475569;
  font-size: 14px;
}

.connection-panel__pill {
  padding: 8px 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid transparent;
}

.connection-panel__pill.is-connected {
  color: #0f766e;
  background: rgba(45, 212, 191, 0.16);
  border-color: rgba(13, 148, 136, 0.32);
}

.connection-panel__pill.is-reconnecting {
  color: #b45309;
  background: rgba(249, 115, 22, 0.18);
  border-color: rgba(249, 115, 22, 0.32);
}

.connection-panel__pill.is-degraded {
  color: #b91c1c;
  background: rgba(248, 113, 113, 0.18);
  border-color: rgba(248, 113, 113, 0.32);
}

.connection-panel__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 18px;
  margin: 0;
}

.connection-panel__grid div {
  background: linear-gradient(135deg, rgba(248, 250, 252, 0.96), rgba(255, 255, 255, 0.98));
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

dt {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #64748b;
}

dd {
  margin: 0;
  font-size: 15px;
  color: #0f172a;
  word-break: break-all;
}

.connection-panel__footer {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: center;
  padding-top: 12px;
  border-top: 1px dashed rgba(148, 163, 184, 0.28);
}

.connection-panel__footer strong {
  display: block;
  font-size: 15px;
  color: #0f172a;
}

.connection-panel__footer p {
  margin: 6px 0 0;
  font-size: 14px;
  color: #475569;
}

.connection-panel__footer button {
  padding: 10px 18px;
  border-radius: 12px;
  border: 1px solid rgba(15, 23, 42, 0.2);
  background: rgba(15, 23, 42, 0.08);
  color: #0f172a;
  font-weight: 600;
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.connection-panel__footer button:hover {
  transform: translateY(-2px);
  background: rgba(15, 23, 42, 0.12);
  box-shadow: 0 12px 22px -18px rgba(15, 23, 42, 0.5);
}

@media (max-width: 768px) {
  .connection-panel {
    padding: 22px;
  }

  .connection-panel__footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
