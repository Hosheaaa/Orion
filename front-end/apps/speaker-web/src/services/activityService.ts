import { httpClient } from "./httpClient";

export interface ActivityDto {
  id: string;
  title: string;
  description: string;
  speaker: string;
  startTime: string;
  endTime?: string | null;
  inputLanguage: string;
  targetLanguages: string[];
  coverUrl?: string | null;
  status: "draft" | "published" | "closed";
  viewerUrl?: string | null;
  createdAt: string;
  updatedAt: string;
}

export async function fetchActivities() {
  const data = await httpClient.get<ActivityDto[]>("/api/v1/activities");
  return data;
}

export async function fetchActivity(id: string) {
  const data = await httpClient.get<ActivityDto>(`/api/v1/activities/${id}`);
  return data;
}
