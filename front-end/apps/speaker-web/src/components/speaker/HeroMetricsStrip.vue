<template>
  <section class="hero-strip" aria-label="实时表现概览">
    <header class="hero-strip__header">
      <div>
        <h1>专注开场，系统已准备就绪</h1>
        <p>
          演讲者控制台实时跟踪观众规模、翻译延迟与字幕准确率，并自动提示语速与术语处理建议，帮助你专注内容本身。
        </p>
      </div>
      <button class="hero-strip__action" type="button">
        <span>下载术语表</span>
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M12 5v11m0 0l-4-4m4 4 4-4M5 19h14"
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.8"
          />
        </svg>
      </button>
    </header>
    <div class="hero-strip__grid">
      <article
        v-for="metric in metrics"
        :key="metric.label"
        class="metric-card"
        :style="{ '--accent': metric.accent }"
      >
        <div class="metric-card__icon" aria-hidden="true">
          <svg viewBox="0 0 32 32">
            <circle cx="16" cy="16" r="14" stroke="var(--accent)" stroke-width="2" fill="none" />
            <path
              v-if="metric.trend === 'up'"
              d="M9 18.5 14.5 12l4 4 5-5"
              stroke="var(--accent)"
              stroke-width="2.4"
              stroke-linecap="round"
              stroke-linejoin="round"
              fill="none"
            />
            <path
              v-else-if="metric.trend === 'down'"
              d="M9 13.5 14.5 20l4-4 5 5"
              stroke="var(--accent)"
              stroke-width="2.4"
              stroke-linecap="round"
              stroke-linejoin="round"
              fill="none"
            />
            <path
              v-else
              d="M9 16h14"
              stroke="var(--accent)"
              stroke-width="2.4"
              stroke-linecap="round"
              fill="none"
            />
          </svg>
        </div>
        <div class="metric-card__body">
          <header>
            <h2>{{ metric.label }}</h2>
            <span class="metric-card__delta">{{ metric.deltaText }}</span>
          </header>
          <strong>{{ metric.value }}</strong>
          <p>{{ metric.description }}</p>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { HeroInsight } from "@/services/mockSpeakerService";

defineProps<{
  metrics: HeroInsight[];
}>();
</script>

<style scoped>
.hero-strip {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 24px;
  padding: 32px 36px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  box-shadow:
    0 28px 56px -36px rgba(15, 23, 42, 0.45),
    0 2px 6px rgba(148, 163, 184, 0.12);
  display: flex;
  flex-direction: column;
  gap: 28px;
  position: relative;
  overflow: hidden;
}

.hero-strip::after {
  content: "";
  position: absolute;
  inset: -40% 10% auto auto;
  width: 320px;
  height: 320px;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.15), transparent 60%);
  filter: blur(110px);
  pointer-events: none;
}

.hero-strip__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
}

.hero-strip__header h1 {
  margin: 0 0 8px;
  font-size: 26px;
  line-height: 1.3;
  color: #0f172a;
}

.hero-strip__header p {
  margin: 0;
  max-width: 520px;
  font-size: 15px;
  color: #475569;
}

.hero-strip__action {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 18px;
  border-radius: 14px;
  border: 1px solid rgba(16, 185, 129, 0.4);
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
  font-weight: 600;
  letter-spacing: 0.01em;
  transition:
    transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.25s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.hero-strip__action svg {
  width: 18px;
  height: 18px;
}

.hero-strip__action:hover {
  transform: translateY(-3px);
  background: rgba(16, 185, 129, 0.16);
  box-shadow:
    0 16px 32px -22px rgba(4, 120, 87, 0.65),
    0 1px 0 rgba(255, 255, 255, 0.8);
}

.hero-strip__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
}

.metric-card {
  position: relative;
  border-radius: 20px;
  overflow: hidden;
  padding: 20px 22px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.95), rgba(249, 250, 252, 0.9));
  border: 1px solid color-mix(in srgb, var(--accent) 22%, transparent);
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 18px;
  align-items: center;
  transition:
    transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.metric-card::before {
  content: "";
  position: absolute;
  inset: -50% 60% 50% -30%;
  background: radial-gradient(circle, color-mix(in srgb, var(--accent) 30%, transparent), transparent 65%);
  opacity: 0.5;
  filter: blur(110px);
}

.metric-card:hover {
  transform: translateY(-4px);
  box-shadow:
    0 28px 40px -40px rgba(15, 23, 42, 0.75),
    0 12px 24px -18px color-mix(in srgb, var(--accent) 35%, transparent);
}

.metric-card__icon {
  width: 54px;
  height: 54px;
  border-radius: 18px;
  background: linear-gradient(135deg, color-mix(in srgb, var(--accent) 16%, transparent), rgba(255, 255, 255, 0.92));
  display: grid;
  place-items: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.metric-card__icon svg {
  width: 36px;
  height: 36px;
}

.metric-card__body header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.metric-card__body h2 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}

.metric-card__delta {
  font-size: 13px;
  color: color-mix(in srgb, var(--accent) 68%, #0f172a);
  background: color-mix(in srgb, var(--accent) 18%, rgba(255, 255, 255, 0.4));
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--accent) 28%, transparent);
}

.metric-card__body strong {
  display: block;
  font-size: 28px;
  margin: 10px 0 6px;
  color: #0f172a;
  letter-spacing: 0.01em;
}

.metric-card__body p {
  margin: 0;
  color: #475569;
  font-size: 14px;
}

@media (max-width: 768px) {
  .hero-strip {
    padding: 24px;
  }

  .hero-strip__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-strip__action {
    width: 100%;
    justify-content: center;
  }

  .metric-card {
    grid-template-columns: 1fr;
    text-align: left;
  }

  .metric-card__icon {
    justify-self: flex-start;
  }
}
</style>
