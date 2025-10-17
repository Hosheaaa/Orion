import type { ActivitySummary, SubtitleItem } from "@/stores/speakerSession";
import { fetchActivities } from "./activityService";
import { httpClient } from "./httpClient";

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

function mapLanguage(code: string) {
  return LANGUAGE_NAME_MAP[code] ?? code;
}

export interface ConsoleActivity extends ActivitySummary {
  displayLanguages: string[];
}

export interface SpeakerTokenResponse {
  token: string;
  expiresAt: string;
}

export interface HeroInsight {
  label: string;
  value: string;
  trend: "up" | "down" | "stable";
  deltaText: string;
  description: string;
  accent: string;
}

export interface GuidanceChecklistItem {
  title: string;
  detail: string;
  emphasis: "primary" | "success" | "warning";
}

/**
 * 拉取活动列表并转换为控制台可用的结构。
 */
export async function fetchConsoleActivities(): Promise<ConsoleActivity[]> {
  const data = await fetchActivities();
  return data.map((item) => ({
    id: item.id,
    title: item.title,
    speaker: item.speaker,
    startTime: item.startTime,
    status: item.status,
    inputLanguage: item.inputLanguage,
    targetLanguages: item.targetLanguages,
    description: item.description,
    viewerUrl: item.viewerUrl ?? undefined,
    displayLanguages: item.targetLanguages.map(mapLanguage)
  }));
}

export async function generateSpeakerToken(activityId: string) {
  const data = await httpClient.post<SpeakerTokenResponse>(
    `/api/v1/activities/${activityId}/tokens/speaker`
  );
  return data;
}

export async function fetchHeroInsights() {
  return httpClient.get<HeroInsight[]>(`/api/v1/speaker-console/hero-insights`);
}

export async function fetchGuidanceChecklist() {
  return httpClient.get<GuidanceChecklistItem[]>(`/api/v1/speaker-console/guidance`);
}

export async function fetchSubtitleHistory(activityId?: string) {
  return httpClient.get<SubtitleItem[]>(`/api/v1/speaker-console/subtitle-history`, {
    query: activityId ? { activityId } : undefined
  });
}
