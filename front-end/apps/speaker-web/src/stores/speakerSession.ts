import { defineStore } from "pinia";
import { computed, ref } from "vue";

export interface ActivitySummary {
  id: string;
  title: string;
  scheduledAt: string;
  venue: string;
  expectedAudience: number;
  translationLanguages: string[];
  description: string;
}

export interface ConnectionSnapshot {
  websocketUrl: string;
  latencyMs: number;
  packetLossRate: number;
  reconnectAttempts: number;
  lastHeartbeatAt: string;
  status: "connected" | "reconnecting" | "degraded";
}

export interface SubtitleItem {
  id: string;
  original: string;
  translated: string;
  timestamp: string;
}

export const useSpeakerSessionStore = defineStore("speakerSession", () => {
  const currentActivity = ref<ActivitySummary | null>(null);
  const isStreaming = ref(false);
  const micLevel = ref(0);
  const connection = ref<ConnectionSnapshot | null>(null);
  const subtitles = ref<SubtitleItem[]>([]);

  const speakableLanguages = computed(() => currentActivity.value?.translationLanguages ?? []);

  function selectActivity(activity: ActivitySummary) {
    currentActivity.value = activity;
  }

  function startStreaming() {
    isStreaming.value = true;
  }

  function stopStreaming() {
    isStreaming.value = false;
  }

  function updateMicLevel(level: number) {
    micLevel.value = Math.max(0, Math.min(level, 1));
  }

  function setConnectionSnapshot(snapshot: ConnectionSnapshot) {
    connection.value = snapshot;
  }

  function pushSubtitle(item: SubtitleItem) {
    subtitles.value = [item, ...subtitles.value].slice(0, 12);
  }

  return {
    currentActivity,
    connection,
    isStreaming,
    micLevel,
    subtitles,
    speakableLanguages,
    selectActivity,
    startStreaming,
    stopStreaming,
    updateMicLevel,
    setConnectionSnapshot,
    pushSubtitle
  };
});
