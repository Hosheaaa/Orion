import type {
  ActivitySummary,
  ConnectionSnapshot,
  SubtitleItem
} from "@/stores/speakerSession";

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

export async function fetchTodayActivities(): Promise<ActivitySummary[]> {
  await wait(220);
  return [
    {
      id: "activity-neo23",
      title: "Orion 全球创新发布会",
      scheduledAt: "2024-08-10T09:30:00+08:00",
      venue: "上海临港•未来剧院",
      expectedAudience: 1200,
      translationLanguages: ["英语", "日语", "西班牙语"],
      description:
        "围绕实时语音翻译技术的落地实践，展示多语言会议在硬件、软件与运营层面的整体方案，并发布最新的 Orion Stream 套件。"
    },
    {
      id: "activity-techleaders",
      title: "科技领袖私享会·午后专场",
      scheduledAt: "2024-08-10T14:30:00+08:00",
      venue: "上海静安•玖和府",
      expectedAudience: 85,
      translationLanguages: ["英语", "法语"],
      description:
        "闭门圆桌形式，聚焦跨国团队协作场景的实时翻译挑战，邀请合作伙伴分享不同市场的落地经验与风险控制策略。"
    },
    {
      id: "activity-media",
      title: "亚太媒体答疑会",
      scheduledAt: "2024-08-11T10:00:00+08:00",
      venue: "线上直播",
      expectedAudience: 450,
      translationLanguages: ["英语", "韩语", "泰语"],
      description:
        "集中回应媒体提问，发布新版本数据安全白皮书，强调实时字幕对资讯透明度的提升，并公布区域化运营计划。"
    }
  ];
}

export async function fetchHeroInsights(): Promise<HeroInsight[]> {
  await wait(180);
  return [
    {
      label: "实时观众",
      value: "1,184",
      trend: "up",
      deltaText: "+12.8% 较上场",
      description: "实时在线观众保持稳步增长，移动端占比 73%。",
      accent: "#10b981"
    },
    {
      label: "翻译延迟",
      value: "1.36s",
      trend: "down",
      deltaText: "-0.4s 再优化",
      description: "端到端延迟保持在 1.5 秒内，满足现场互动需求。",
      accent: "#8b5cf6"
    },
    {
      label: "字幕准确率",
      value: "97.2%",
      trend: "stable",
      deltaText: "稳定在 ≥97%",
      description: "结合术语表与人工校对流程，保障品牌关键词准确呈现。",
      accent: "#f97316"
    }
  ];
}

export async function fetchConnectionSnapshot(): Promise<ConnectionSnapshot> {
  await wait(160);
  return {
    websocketUrl: "wss://orion-demo.live/ws/speaker?activity=activity-neo23",
    latencyMs: 84,
    packetLossRate: 0.6,
    reconnectAttempts: 0,
    lastHeartbeatAt: new Date().toISOString(),
    status: "connected"
  };
}

export async function fetchSubtitleHistory(): Promise<SubtitleItem[]> {
  await wait(200);
  return [
    {
      id: "sub-14319",
      original: "大家下午好，欢迎来到 Orion 全球创新发布会的现场。",
      translated: "Good afternoon. Welcome to the Orion Global Innovation Launch.",
      timestamp: new Date(Date.now() - 25_000).toISOString()
    },
    {
      id: "sub-14320",
      original: "今天我们将公布全新的实时语音翻译引擎。",
      translated: "Today we unveil our next-generation real-time speech translation engine.",
      timestamp: new Date(Date.now() - 19_000).toISOString()
    },
    {
      id: "sub-14321",
      original: "它支持 36 种语言的同步字幕，并具备上下文学习能力。",
      translated: "It delivers synchronized captions in 36 languages with contextual learning.",
      timestamp: new Date(Date.now() - 12_000).toISOString()
    }
  ];
}

export async function fetchGuidanceChecklist(): Promise<GuidanceChecklistItem[]> {
  await wait(140);
  return [
    {
      title: "语速控制在 150 字 / 分钟以内",
      detail: "确保 STT 引擎保持 ≥96% 准确率，并让译员有缓冲空间。",
      emphasis: "primary"
    },
    {
      title: "段落之间预留 1.5 秒停顿",
      detail: "便于翻译引擎分段处理，降低字幕延迟并提升滚动体验。",
      emphasis: "success"
    },
    {
      title: "遇到专有名词先拼写后阐述",
      detail: "帮助实时术语库准确抓取，避免字幕出现模糊词汇。",
      emphasis: "warning"
    }
  ];
}

function wait(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
