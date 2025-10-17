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
  // 部分后端实现返回 null 会导致前端后续 map 触发异常，这里做守护处理
  return Array.isArray(data) ? data : [];
}

export async function fetchActivity(id: string) {
  const data = await httpClient.get<ActivityDto>(`/api/v1/activities/${id}`);
  return data;
}

export interface CreateActivityPayload {
  title: string;
  description: string;
  speaker: string;
  startTime: string;
  inputLanguage: string;
  targetLanguages: string[];
  coverUrl?: string | null;
}

export interface UpdateActivityPayload {
  title?: string;
  description?: string;
  speaker?: string;
  startTime?: string;
  inputLanguage?: string;
  targetLanguages?: string[];
  coverUrl?: string | null;
}

export async function createActivity(payload: CreateActivityPayload) {
  const data = await httpClient.post<ActivityDto>("/api/v1/activities", payload);
  return data;
}

export async function updateActivity(id: string, payload: UpdateActivityPayload) {
  const data = await httpClient.put<ActivityDto>(`/api/v1/activities/${id}`, payload);
  return data;
}

export async function publishActivity(id: string) {
  const data = await httpClient.post<ActivityDto>(`/api/v1/activities/${id}/publish`);
  return data;
}

export async function closeActivity(id: string) {
  const data = await httpClient.post<ActivityDto>(`/api/v1/activities/${id}/close`);
  return data;
}

export async function deleteActivity(id: string) {
  await httpClient.delete(`/api/v1/activities/${id}`);
}
