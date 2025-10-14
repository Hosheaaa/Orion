import { useQuery } from "@tanstack/vue-query";
import {
  fetchHeroInsights,
  fetchTodayActivities,
  fetchConnectionSnapshot,
  fetchSubtitleHistory,
  fetchGuidanceChecklist
} from "@/services/mockSpeakerService";

export function useSpeakerActivities() {
  return useQuery({
    queryKey: ["speaker", "activities"],
    queryFn: fetchTodayActivities,
    staleTime: 1000 * 60
  });
}

export function useHeroInsights() {
  return useQuery({
    queryKey: ["speaker", "hero-insights"],
    queryFn: fetchHeroInsights
  });
}

export function useConnectionSnapshot() {
  return useQuery({
    queryKey: ["speaker", "connection"],
    queryFn: fetchConnectionSnapshot,
    refetchInterval: 1000 * 20
  });
}

export function useSubtitleHistory() {
  return useQuery({
    queryKey: ["speaker", "subtitle-history"],
    queryFn: fetchSubtitleHistory,
    refetchInterval: 1000 * 30
  });
}

export function useGuidanceChecklist() {
  return useQuery({
    queryKey: ["speaker", "guidance"],
    queryFn: fetchGuidanceChecklist,
    staleTime: 1000 * 60 * 5
  });
}
