const STORAGE_KEY = "orion:speaker:manual-tokens";

export interface ManualSpeakerTokenRecord {
  token: string;
  updatedAt: string;
}

function loadAll(): Record<string, ManualSpeakerTokenRecord> {
  if (typeof window === "undefined") {
    return {};
  }
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw) as Record<string, ManualSpeakerTokenRecord>;
    if (parsed && typeof parsed === "object") {
      return parsed;
    }
    return {};
  } catch {
    return {};
  }
}

function saveAll(records: Record<string, ManualSpeakerTokenRecord>) {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(records));
}

export function loadSpeakerToken(activityId: string): ManualSpeakerTokenRecord | null {
  const records = loadAll();
  return records[activityId] ?? null;
}

export function saveSpeakerToken(activityId: string, record: ManualSpeakerTokenRecord) {
  const records = loadAll();
  records[activityId] = {
    token: record.token,
    updatedAt: record.updatedAt
  };
  saveAll(records);
}

export function clearSpeakerToken(activityId: string) {
  const records = loadAll();
  if (records[activityId]) {
    delete records[activityId];
    saveAll(records);
  }
}

export function clearAllSpeakerTokens() {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.removeItem(STORAGE_KEY);
}
