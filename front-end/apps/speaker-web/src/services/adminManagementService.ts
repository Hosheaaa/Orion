import { httpClient } from "./httpClient";

export type TokenStatus = "active" | "revoked" | "expired";
export type TokenType = "speaker" | "viewer";

export interface ActivityTokenRecord {
  id: string;
  type: TokenType;
  value: string;
  status: TokenStatus;
  createdAt: string;
  expiresAt: string;
  maxAudience?: number;
}

export interface ViewerEntryResponse {
  activityId: string;
  shareUrl: string;
  qrType: string;
  qrContent: string;
  status: "inactive" | "active" | "revoked";
  updatedAt: string;
}

export interface GenerateViewerTokenPayload {
  ttlMinutes?: number;
  maxAudience?: number;
}

export interface GenerateViewerTokenResponse {
  code: string;
  expiresAt: string;
}

export async function listActivityTokens(activityId: string) {
  return httpClient.get<ActivityTokenRecord[]>(`/api/v1/activities/${activityId}/tokens`);
}

export async function generateViewerToken(activityId: string, payload: GenerateViewerTokenPayload = {}) {
  return httpClient.post<GenerateViewerTokenResponse>(
    `/api/v1/activities/${activityId}/tokens/viewer`,
    payload
  );
}

export async function revokeSpeakerTokens(activityId: string) {
  return httpClient.post<void>(`/api/v1/activities/${activityId}/tokens/speaker/revoke`);
}

export async function revokeSpeakerToken(activityId: string, tokenId: string) {
  return httpClient.post<void>(
    `/api/v1/activities/${activityId}/tokens/speaker/${tokenId}/revoke`
  );
}

export async function getViewerEntry(activityId: string) {
  return httpClient.get<ViewerEntryResponse>(`/api/v1/activities/${activityId}/viewer-entry`);
}

export async function revokeViewerEntry(activityId: string) {
  return httpClient.post<ViewerEntryResponse>(
    `/api/v1/activities/${activityId}/viewer-entry/revoke`
  );
}

export async function activateViewerEntry(activityId: string) {
  return httpClient.post<ViewerEntryResponse>(
    `/api/v1/activities/${activityId}/viewer-entry/activate`
  );
}
