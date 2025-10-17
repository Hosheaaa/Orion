import type { ActivitySummary } from "@/stores/speakerSession";
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
