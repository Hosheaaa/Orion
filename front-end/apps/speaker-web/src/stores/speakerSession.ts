import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { generateSpeakerToken } from "@/services/speakerConsoleService";
import { envConfig } from "@/config/env";
import {
  calculateRMS,
  downsampleBuffer,
  floatTo16BitPCM,
  pcm16ToBase64
} from "@/services/audioUtils";

export type ActivityStatus = "draft" | "published" | "closed";

export interface ActivitySummary {
  id: string;
  title: string;
  speaker: string;
  startTime: string;
  status: ActivityStatus;
  inputLanguage: string;
  targetLanguages: string[];
  description: string;
  viewerUrl?: string;
}

export interface ConnectionSnapshot {
  websocketUrl: string;
  latencyMs: number;
  packetLossRate: number;
  reconnectAttempts: number;
  lastHeartbeatAt: string;
  status: "idle" | "connected" | "reconnecting" | "degraded";
  stateMessage?: string;
}

export interface SubtitleItem {
  id: string;
  original: string;
  translated: string;
  timestamp: string;
}

type StreamingStatus = "idle" | "connecting" | "streaming";

export const useSpeakerSessionStore = defineStore("speakerSession", () => {
  const currentActivity = ref<ActivitySummary | null>(null);
  const streamingStatus = ref<StreamingStatus>("idle");
  const micLevel = ref(0);
  const connection = ref<ConnectionSnapshot | null>(null);
  const subtitles = ref<SubtitleItem[]>([]);
  const speakerToken = ref<string | null>(null);
  const speakerTokenExpiresAt = ref<string | null>(null);
  const lastError = ref<string | null>(null);

  const websocket = ref<WebSocket | null>(null);
  let mediaStream: MediaStream | null = null;
  let audioContext: AudioContext | null = null;
  let sourceNode: MediaStreamAudioSourceNode | null = null;
  let processorNode: ScriptProcessorNode | null = null;
  let muteGain: GainNode | null = null;

  const sequence = ref(0);
  const reconnectAttempts = ref(0);

  const speakableLanguages = computed(() => currentActivity.value?.targetLanguages ?? []);
  const isStreaming = computed(() => streamingStatus.value === "streaming");

  const defaultConnectionSnapshot = () => ({
    websocketUrl: "",
    latencyMs: 0,
    packetLossRate: 0,
    reconnectAttempts: 0,
    lastHeartbeatAt: new Date().toISOString(),
    status: "idle" as const,
    stateMessage: ""
  });

  function selectActivity(activity: ActivitySummary) {
    currentActivity.value = activity;
    speakerToken.value = null;
    speakerTokenExpiresAt.value = null;
    subtitles.value = [];
    if (connection.value?.status === "connected") {
      connection.value = defaultConnectionSnapshot();
    }
  }

  function updateMicLevel(level: number) {
    micLevel.value = Math.max(0, Math.min(level, 1));
  }

  function updateConnection(partial: Partial<ConnectionSnapshot>) {
    const base = connection.value ?? defaultConnectionSnapshot();
    connection.value = {
      ...base,
      ...partial,
      lastHeartbeatAt: partial.lastHeartbeatAt ?? new Date().toISOString()
    };
  }

  async function ensureSpeakerToken() {
    const activity = currentActivity.value;
    if (!activity) return null;

    if (speakerToken.value && speakerTokenExpiresAt.value) {
      const expiresAt = new Date(speakerTokenExpiresAt.value).getTime();
      if (expiresAt - Date.now() > 60_000) {
        return speakerToken.value;
      }
    }

    const data = await generateSpeakerToken(activity.id);
    speakerToken.value = data.token;
    speakerTokenExpiresAt.value = data.expiresAt;
    return data.token;
  }

  async function startStreaming() {
    if (streamingStatus.value === "streaming" || streamingStatus.value === "connecting") {
      return;
    }

    const activity = currentActivity.value;
    if (!activity) {
      throw new Error("请先选择活动后再开始推流。");
    }

    streamingStatus.value = "connecting";
    lastError.value = null;

    try {
      const token = await ensureSpeakerToken();
      if (!token) {
        throw new Error("未能获取到有效的推流令牌。");
      }

      mediaStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 48000,
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      });

      audioContext = new AudioContext({ sampleRate: 48000 });
      await audioContext.resume();

      sourceNode = audioContext.createMediaStreamSource(mediaStream);
      processorNode = audioContext.createScriptProcessor(4096, 1, 1);
      muteGain = audioContext.createGain();
      muteGain.gain.value = 0;

      sourceNode.connect(processorNode);
      processorNode.connect(muteGain);
      muteGain.connect(audioContext.destination);

      processorNode.onaudioprocess = (event) => {
        const channelData = event.inputBuffer.getChannelData(0);
        handleAudioFrame(channelData, audioContext?.sampleRate ?? 48000);
      };

      const wsUrl = buildSpeakerWsUrl(activity.id, token, activity.inputLanguage);
      sequence.value = 0;
      reconnectAttempts.value = 0;
      connection.value = {
        websocketUrl: wsUrl,
        latencyMs: 0,
        packetLossRate: 0,
        reconnectAttempts: reconnectAttempts.value,
        lastHeartbeatAt: new Date().toISOString(),
        status: "reconnecting",
        stateMessage: "正在建立 WebSocket 连接..."
      };

      await establishWebSocket(wsUrl);
      streamingStatus.value = "streaming";
      updateConnection({
        status: "connected",
        stateMessage: "推流连接成功"
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      lastError.value = message;
      await stopStreaming(true);
      streamingStatus.value = "idle";
      updateConnection({
        status: "degraded",
        stateMessage: message
      });
      throw error;
    }
  }

  async function stopStreaming(silent = false) {
    if (streamingStatus.value === "idle" && !websocket.value) {
      return;
    }

    const ws = websocket.value;
    if (ws && ws.readyState === WebSocket.OPEN) {
      sendControl(ws, "STOP");
    }

    cleanupStreamingResources();

    streamingStatus.value = "idle";
    updateMicLevel(0);

    if (!silent) {
      updateConnection({
        status: "idle",
        stateMessage: "推流已停止"
      });
    }
  }

  async function establishWebSocket(url: string) {
    return new Promise<void>((resolve, reject) => {
      const ws = new WebSocket(url);
      websocket.value = ws;

      const handleOpen = () => {
        ws.removeEventListener("error", handleError);
        ws.send(
          JSON.stringify({
            type: "CONTROL",
            payload: { action: "START" }
          })
        );
        resolve();
      };

      const handleError = (event: Event) => {
        ws.removeEventListener("open", handleOpen);
        reject(new Error("WebSocket 连接失败，请检查后端服务或网络状态。"));
      };

      ws.addEventListener("open", handleOpen, { once: true });
      ws.addEventListener("error", handleError, { once: true });

      ws.onmessage = (event) => handleSocketMessage(event.data);
      ws.onclose = () => {
        if (streamingStatus.value === "streaming") {
          updateConnection({
            status: "degraded",
            stateMessage: "连接已断开"
          });
        }
        cleanupStreamingResources();
        streamingStatus.value = "idle";
      };
      ws.onerror = () => {
        updateConnection({
          status: "degraded",
          stateMessage: "WebSocket 出现异常"
        });
      };
    });
  }

  function handleSocketMessage(raw: string) {
    let parsed: { type: string; payload?: any };
    try {
      parsed = JSON.parse(raw);
    } catch (error) {
      console.warn("无法解析 WebSocket 消息", error);
      return;
    }

    switch (parsed.type) {
      case "STATE": {
        const status = String(parsed.payload?.status ?? "").toUpperCase();
        const message = parsed.payload?.message ?? "";
        updateConnection({
          status: mapStateToSnapshot(status),
          stateMessage: message
        });
        if (status === "STOPPED") {
          streamingStatus.value = "idle";
        }
        break;
      }
      case "ERROR": {
        const message = parsed.payload?.message ?? "推流出现未知错误";
        lastError.value = message;
        updateConnection({
          status: "degraded",
          stateMessage: message
        });
        streamingStatus.value = "idle";
        break;
      }
      case "SUBTITLE": {
        const payload = parsed.payload ?? {};
        const subtitle: SubtitleItem = {
          id: payload.id ?? crypto.randomUUID(),
          original: payload.original ?? "",
          translated: payload.text ?? payload.original ?? "",
          timestamp: payload.timestamp ?? new Date().toISOString()
        };
        pushSubtitle(subtitle);
        break;
      }
      default:
        break;
    }
  }

  function handleAudioFrame(channelData: Float32Array, sampleRate: number) {
    const ws = websocket.value;
    if (!ws || ws.readyState !== WebSocket.OPEN || streamingStatus.value !== "streaming") {
      return;
    }

    const downsampled = downsampleBuffer(channelData, sampleRate, 16000);
    if (!downsampled.length) {
      return;
    }

    const pcm16 = floatTo16BitPCM(downsampled);
    const base64 = pcm16ToBase64(pcm16);
    if (!base64) {
      return;
    }

    sequence.value += 1;
    ws.send(
      JSON.stringify({
        type: "AUDIO",
        payload: {
          chunk: base64,
          sequence: sequence.value
        }
      })
    );

    const rms = calculateRMS(downsampled);
    updateMicLevel(Math.min(1, rms * 8));
    updateConnection({
      lastHeartbeatAt: new Date().toISOString()
    });
  }

  function buildSpeakerWsUrl(activityId: string, token: string, language: string) {
    const base = envConfig.wsBaseUrl.replace(/\/$/, "");
    const params = new URLSearchParams({
      activityId,
      token,
      language
    });
    return `${base}/ws/speaker?${params.toString()}`;
  }

  function mapStateToSnapshot(state: string): ConnectionSnapshot["status"] {
    switch (state) {
      case "READY":
      case "STREAMING":
        return "connected";
      case "STOPPED":
        return "idle";
      default:
        return "degraded";
    }
  }

  function sendControl(ws: WebSocket, action: "START" | "STOP") {
    if (ws.readyState !== WebSocket.OPEN) return;
    ws.send(
      JSON.stringify({
        type: "CONTROL",
        payload: { action }
      })
    );
  }

  function cleanupStreamingResources() {
    if (processorNode) {
      processorNode.disconnect();
      processorNode.onaudioprocess = null;
      processorNode = null;
    }
    if (sourceNode) {
      sourceNode.disconnect();
      sourceNode = null;
    }
    if (muteGain) {
      muteGain.disconnect();
      muteGain = null;
    }
    if (audioContext) {
      audioContext.close().catch(() => {});
      audioContext = null;
    }
    if (mediaStream) {
      mediaStream.getTracks().forEach((track) => track.stop());
      mediaStream = null;
    }

    if (websocket.value) {
      const ws = websocket.value;
      websocket.value = null;
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close(1000, "client close");
      }
    }

    updateMicLevel(0);
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
    ensureSpeakerToken,
    updateMicLevel,
    pushSubtitle,
    speakerToken,
    speakerTokenExpiresAt,
    streamingStatus,
    lastError
  };
});
