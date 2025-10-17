<template>
  <div class="admin-console">
    <section class="admin-console__hero">
      <div>
        <h1>活动管理后台</h1>
        <p>
          管理活动排期、演讲者令牌与观众入口。所有改动即时生效，请在正式发布前完成彩排与权限校验。
        </p>
      </div>
      <button class="hero__action" type="button" @click="openCreateModal">
        新建活动
      </button>
    </section>

    <section class="admin-console__stats">
      <article class="stat-card">
        <header>全部活动</header>
        <strong>{{ totalActivities }}</strong>
        <span>含草稿 / 已发布 / 已关闭</span>
      </article>
      <article class="stat-card">
        <header>已发布</header>
        <strong>{{ publishedCount }}</strong>
        <span>当前可供演讲者与观众接入</span>
      </article>
      <article class="stat-card">
        <header>即将开始</header>
        <strong>{{ upcomingCount }}</strong>
        <span>起始时间在未来 24 小时内</span>
      </article>
    </section>

    <section class="admin-console__grid">
      <div class="admin-console__primary">
        <div class="admin-panel">
          <header class="admin-panel__header">
            <div>
              <h2>活动列表</h2>
              <p>选择一场活动以查看详情、生成令牌或调整发布状态。</p>
            </div>
            <div class="filter-group">
              <button
                v-for="status in statusFilters"
                :key="status.value"
                type="button"
                class="filter-chip"
                :class="{ 'is-active': statusFilter === status.value }"
                @click="statusFilter = status.value"
              >
                {{ status.label }}
              </button>
            </div>
          </header>

          <div class="activity-list" v-if="activitiesLoading">
            <p class="muted">正在加载活动列表...</p>
          </div>
          <div class="activity-list" v-else-if="!filteredActivities.length">
            <p class="muted">
              {{ statusFilter === "all" ? "暂无活动，请先创建一场新活动。" : "当前筛选条件下没有活动。" }}
            </p>
          </div>
          <div class="activity-list" v-else>
            <article
              v-for="activity in filteredActivities"
              :key="activity.id"
              :class="['activity-card', { 'is-active': activity.id === selectedActivityId }]"
              role="button"
              tabindex="0"
              @click="selectActivity(activity.id)"
              @keyup.enter="selectActivity(activity.id)"
            >
              <header>
                <div>
                  <h3>{{ activity.title }}</h3>
                  <span class="activity-meta">
                    {{ formatDateTime(activity.startTime) }} · {{ formatStatus(activity.status) }} · {{ activity.speaker }}
                  </span>
                </div>
                <span class="badge">输入语种：{{ formatLanguage(activity.inputLanguage) }}</span>
              </header>
              <p class="activity-description">{{ activity.description || "尚未填写活动简介" }}</p>
              <footer>
                <div class="lang-chips">
                  <span v-for="lang in activity.targetLanguages" :key="lang" class="lang-chip">
                    {{ formatLanguage(lang) }}
                  </span>
                </div>
                <div class="card-cta">
                  <span>查看详情</span>
                  <svg viewBox="0 0 20 20" aria-hidden="true">
                    <path
                      d="M7.5 4.5 12.5 10 7.5 15.5"
                      fill="none"
                      stroke="currentColor"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.6"
                    />
                  </svg>
                </div>
              </footer>
            </article>
          </div>
        </div>
      </div>

      <aside class="admin-console__secondary">
        <div class="admin-panel" v-if="selectedActivity">
          <header class="admin-panel__header">
            <div>
              <h2>{{ selectedActivity.title }}</h2>
              <p>活动详情与生命周期管理</p>
            </div>
            <span class="status-badge" :data-status="selectedActivity.status">
              {{ formatStatus(selectedActivity.status) }}
            </span>
          </header>

          <dl class="detail-grid">
            <div>
              <dt>主讲人</dt>
              <dd>{{ selectedActivity.speaker }}</dd>
            </div>
            <div>
              <dt>开始时间</dt>
              <dd>{{ formatDateTime(selectedActivity.startTime) }}</dd>
            </div>
            <div>
              <dt>创建时间</dt>
              <dd>{{ formatDateTime(selectedActivity.createdAt) }}</dd>
            </div>
            <div>
              <dt>最后更新</dt>
              <dd>{{ formatDateTime(selectedActivity.updatedAt) }}</dd>
            </div>
          </dl>

          <section class="detail-section">
            <h3>目标语言</h3>
            <div class="lang-chips">
              <span v-for="lang in selectedActivity.targetLanguages" :key="lang" class="lang-chip">
                {{ formatLanguage(lang) }}
              </span>
            </div>
          </section>

          <section class="detail-section">
            <h3>观众端链接</h3>
            <div class="link-row" v-if="selectedActivity.viewerUrl">
              <a :href="selectedActivity.viewerUrl" target="_blank" rel="noopener">{{ selectedActivity.viewerUrl }}</a>
              <button type="button" class="ghost" @click="copyToClipboard(selectedActivity.viewerUrl, '观众链接')">复制</button>
            </div>
            <p v-else class="muted">尚未生成观众入口，请完成活动发布与邀请码生成。</p>
          </section>

          <section class="detail-actions">
            <button type="button" class="primary" @click="openEditModal" :disabled="!selectedActivity">
              编辑活动
            </button>
            <button
              type="button"
              class="secondary"
              @click="handlePublish"
              :disabled="selectedActivity.status !== 'draft' || publishPending"
            >
              {{ publishPending ? "发布中..." : "发布活动" }}
            </button>
            <button
              type="button"
              class="secondary"
              @click="handleClose"
              :disabled="selectedActivity.status !== 'published' || closePending"
            >
              {{ closePending ? "关闭中..." : "关闭活动" }}
            </button>
            <button
              type="button"
              class="danger"
              @click="handleDelete"
              :disabled="selectedActivity.status !== 'draft' || deletePending"
            >
              {{ deletePending ? "删除中..." : "删除草稿" }}
            </button>
          </section>
        </div>
        <div class="admin-panel empty" v-else>
          <p class="muted">请选择一场活动以查看详细信息和操作项。</p>
        </div>

        <div class="admin-panel" v-if="selectedActivity">
          <header class="admin-panel__header">
            <div>
              <h2>令牌管理</h2>
              <p>签发演讲者推流令牌与观众邀请码，支持随时撤销。</p>
            </div>
          </header>

          <section class="token-section">
            <div class="token-section__header">
              <h3>演讲者令牌</h3>
              <div class="token-section__actions">
                <button
                  type="button"
                  class="secondary"
                  @click="handleGenerateSpeakerToken"
                  :disabled="speakerTokenPending"
                >
                  {{ speakerTokenPending ? "生成中..." : "生成新令牌" }}
                </button>
                <button
                  type="button"
                  class="danger"
                  @click="handleRevokeSpeakerTokens"
                  :disabled="revokeSpeakerTokensPending || !speakerTokens.length"
                >
                  {{ revokeSpeakerTokensPending ? "撤销中..." : "撤销全部" }}
                </button>
              </div>
            </div>
            <p class="muted token-section__hint">
              令牌仅展示给管理员，默认有效期 24 小时，请通过可信渠道发送给演讲者，演讲者端需手动输入后才能推流。
            </p>
            <div v-if="tokensLoading" class="muted">正在加载令牌...</div>
            <ul v-else-if="speakerTokens.length" class="token-list">
              <li v-for="token in speakerTokens" :key="token.id">
                <div>
                  <strong>{{ token.value }}</strong>
                  <span class="token-meta">状态：{{ formatTokenStatus(token.status) }} · 过期时间 {{ formatDateTime(token.expiresAt) }}</span>
                </div>
                <div class="token-list__actions">
                  <button type="button" class="ghost" @click="copyToClipboard(token.value, '演讲者令牌')">复制</button>
                  <button
                    type="button"
                    class="danger ghost"
                    @click="handleRevokeSingleSpeakerToken(token)"
                    :disabled="revokeSpeakerTokenPending || token.status !== 'active'"
                  >
                    撤销
                  </button>
                </div>
              </li>
            </ul>
            <p v-else class="muted">暂无演讲者令牌，生成后将在此展示。</p>
          </section>

          <section class="token-section">
            <div class="token-section__header">
              <h3>观众邀请码</h3>
              <form class="viewer-token-form" @submit.prevent="handleGenerateViewerToken">
                <label>
                  有效期（分钟）
                  <input
                    type="number"
                    min="10"
                    step="10"
                    v-model.number="viewerTokenForm.ttlMinutes"
                    placeholder="默认 120"
                  />
                </label>
                <label>
                  最大观众数（可选）
                  <input
                    type="number"
                    min="1"
                    v-model.number="viewerTokenForm.maxAudience"
                    placeholder="不填则不限制"
                  />
                </label>
                <button type="submit" class="secondary" :disabled="viewerTokenPending">
                  {{ viewerTokenPending ? "生成中..." : "生成邀请码" }}
                </button>
              </form>
            </div>
            <div v-if="tokensLoading" class="muted">正在加载邀请码...</div>
            <ul v-else-if="viewerTokens.length" class="token-list">
              <li v-for="token in viewerTokens" :key="token.id">
                <div>
                  <strong>{{ token.value }}</strong>
                  <span class="token-meta">
                    状态：{{ formatTokenStatus(token.status) }} · 过期时间 {{ formatDateTime(token.expiresAt) }}
                    <template v-if="token.maxAudience"> · 限额 {{ token.maxAudience }} 人</template>
                  </span>
                </div>
                <button type="button" class="ghost" @click="copyToClipboard(token.value, '观众邀请码')">复制</button>
              </li>
            </ul>
            <p v-else class="muted">尚未生成观众邀请码。</p>
          </section>
        </div>

        <div class="admin-panel" v-if="selectedActivity">
          <header class="admin-panel__header">
            <div>
              <h2>观众入口状态</h2>
              <p>二维码与分享链接状态同步至观众端，请谨慎操作。</p>
            </div>
          </header>
          <div v-if="viewerEntryLoading" class="muted">正在加载入口信息...</div>
          <div v-else-if="viewerEntry" class="viewer-entry">
            <p>
              当前状态：
              <span class="status-badge" :data-status="viewerEntry.status">{{ formatViewerStatus(viewerEntry.status) }}</span>
            </p>
            <p>
              分享链接：
              <a :href="viewerEntry.shareUrl" target="_blank" rel="noopener">{{ viewerEntry.shareUrl }}</a>
            </p>
            <p class="muted">最近更新：{{ formatDateTime(viewerEntry.updatedAt) }}</p>
            <div class="qr-preview" v-if="viewerEntry.qrContent">
              <textarea readonly>{{ viewerEntry.qrContent }}</textarea>
              <button type="button" class="ghost" @click="copyToClipboard(viewerEntry.qrContent, '二维码内容')">
                复制二维码内容
              </button>
            </div>
            <div class="viewer-entry__actions">
              <button
                type="button"
                class="secondary"
                @click="handleActivateEntry"
                :disabled="viewerEntry.status === 'active' || activateEntryPending"
              >
                {{ activateEntryPending ? "处理中..." : "启用入口" }}
              </button>
              <button
                type="button"
                class="danger"
                @click="handleRevokeEntry"
                :disabled="viewerEntry.status !== 'active' || revokeEntryPending"
              >
                {{ revokeEntryPending ? "处理中..." : "撤销入口" }}
              </button>
            </div>
          </div>
          <p v-else class="muted">尚未生成观众入口，生成邀请码后会自动创建。</p>
        </div>
      </aside>
    </section>

    <div v-if="showCreateModal" class="modal">
      <div class="modal__backdrop" @click="closeCreateModal" />
      <div class="modal__content">
        <header>
          <h2>新建活动</h2>
          <p>补充基础信息，稍后可继续编辑。</p>
        </header>
        <form class="modal-form" @submit.prevent="submitCreate">
          <label>
            活动标题
            <n-input v-model:value="createForm.title" placeholder="请输入活动标题" />
          </label>
          <label>
            主讲人
            <n-input v-model:value="createForm.speaker" placeholder="例如：李雷" />
          </label>
          <label>
            开始时间
            <input type="datetime-local" v-model="createForm.startTime" />
          </label>
          <label>
            输入语种
            <select v-model="createForm.inputLanguage">
              <option v-for="lang in languageOptions" :key="lang.code" :value="lang.code">
                {{ lang.label }}
              </option>
            </select>
          </label>
          <label>
            目标语种（至少选择一个）
            <div class="language-selector">
              <button
                v-for="lang in languageOptions"
                :key="lang.code"
                type="button"
                class="language-chip"
                :class="{ 'is-selected': createForm.targetLanguages.includes(lang.code) }"
                @click="toggleLanguage(createForm.targetLanguages, lang.code)"
              >
                {{ lang.label }}
              </button>
            </div>
          </label>
          <label>
            活动简介
            <n-input v-model:value="createForm.description" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
          </label>
          <label>
            封面地址（可选）
            <n-input v-model:value="createForm.coverUrl" placeholder="https://example.com/cover.png" />
          </label>
          <footer class="modal-actions">
            <button type="button" class="ghost" @click="closeCreateModal">取消</button>
            <button type="submit" class="primary" :disabled="createPending">
              {{ createPending ? "创建中..." : "创建活动" }}
            </button>
          </footer>
        </form>
      </div>
    </div>

    <div v-if="showEditModal && selectedActivity" class="modal">
      <div class="modal__backdrop" @click="closeEditModal" />
      <div class="modal__content">
        <header>
          <h2>编辑活动</h2>
          <p>更新活动基本信息，保存后即时生效。</p>
        </header>
        <form class="modal-form" @submit.prevent="submitEdit">
          <label>
            活动标题
            <n-input v-model:value="editForm.title" />
          </label>
          <label>
            主讲人
            <n-input v-model:value="editForm.speaker" />
          </label>
          <label>
            开始时间
            <input type="datetime-local" v-model="editForm.startTime" />
          </label>
          <label>
            输入语种
            <select v-model="editForm.inputLanguage">
              <option v-for="lang in languageOptions" :key="lang.code" :value="lang.code">
                {{ lang.label }}
              </option>
            </select>
          </label>
          <label>
            目标语种
            <div class="language-selector">
              <button
                v-for="lang in languageOptions"
                :key="lang.code"
                type="button"
                class="language-chip"
                :class="{ 'is-selected': editForm.targetLanguages.includes(lang.code) }"
                @click="toggleLanguage(editForm.targetLanguages, lang.code)"
              >
                {{ lang.label }}
              </button>
            </div>
          </label>
          <label>
            活动简介
            <n-input v-model:value="editForm.description" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
          </label>
          <label>
            封面地址
            <n-input v-model:value="editForm.coverUrl" />
          </label>
          <footer class="modal-actions">
            <button type="button" class="ghost" @click="closeEditModal">取消</button>
            <button type="submit" class="primary" :disabled="updatePending">
              {{ updatePending ? "保存中..." : "保存修改" }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, reactive, ref, watch } from "vue";
import { useQuery, useMutation } from "@tanstack/vue-query";
import type { MessageApiInjection } from "naive-ui/es/message/src/MessageProvider";
import {
  fetchActivities,
  createActivity,
  updateActivity,
  publishActivity,
  closeActivity,
  deleteActivity,
  type ActivityDto,
  type CreateActivityPayload,
  type UpdateActivityPayload
} from "@/services/activityService";
import { generateSpeakerToken } from "@/services/speakerConsoleService";
import {
  listActivityTokens,
  generateViewerToken,
  revokeSpeakerTokens,
  revokeSpeakerToken,
  getViewerEntry,
  revokeViewerEntry,
  activateViewerEntry,
  type ActivityTokenRecord,
  type ViewerEntryResponse,
  type GenerateViewerTokenPayload
} from "@/services/adminManagementService";

const message = inject<MessageApiInjection | undefined>("naive-message");

const statusFilters = [
  { label: "全部", value: "all" as const },
  { label: "草稿", value: "draft" as const },
  { label: "已发布", value: "published" as const },
  { label: "已关闭", value: "closed" as const }
];

const languageOptions = [
  { code: "zh-CN", label: "简体中文" },
  { code: "zh-TW", label: "繁体中文" },
  { code: "en", label: "英语" },
  { code: "ja", label: "日语" },
  { code: "ko", label: "韩语" },
  { code: "es", label: "西班牙语" },
  { code: "fr", label: "法语" },
  { code: "de", label: "德语" },
  { code: "it", label: "意大利语" },
  { code: "pt", label: "葡萄牙语" },
  { code: "ru", label: "俄语" }
];

const statusFilter = ref<typeof statusFilters[number]["value"]>("all");

const {
  data: activitiesData,
  isLoading: activitiesLoading,
  refetch: refetchActivities
} = useQuery({
  queryKey: ["admin", "activities"],
  queryFn: fetchActivities,
  staleTime: 1000 * 30
});

const selectedActivityId = ref<string | null>(null);

watch(activitiesData, (list) => {
  if (!list || !list.length) {
    selectedActivityId.value = null;
    return;
  }
  if (!selectedActivityId.value) {
    selectedActivityId.value = list[0].id;
    return;
  }
  const stillExists = list.some((item) => item.id === selectedActivityId.value);
  if (!stillExists) {
    selectedActivityId.value = list[0].id;
  }
});

const filteredActivities = computed(() => {
  const list = activitiesData.value ?? [];
  if (statusFilter.value === "all") {
    return list;
  }
  return list.filter((item) => item.status === statusFilter.value);
});

const selectedActivity = computed<ActivityDto | null>(() => {
  if (!selectedActivityId.value) return null;
  return (activitiesData.value ?? []).find((item) => item.id === selectedActivityId.value) ?? null;
});

watch(filteredActivities, (list) => {
  if (!list.length) {
    return;
  }
  if (!selectedActivityId.value || !list.some((item) => item.id === selectedActivityId.value)) {
    selectedActivityId.value = list[0].id;
  }
});

const totalActivities = computed(() => activitiesData.value?.length ?? 0);
const publishedCount = computed(
  () => (activitiesData.value ?? []).filter((item) => item.status === "published").length
);
const upcomingCount = computed(() => {
  const now = Date.now();
  const threshold = now + 24 * 60 * 60 * 1000;
  return (activitiesData.value ?? []).filter((item) => {
    const time = new Date(item.startTime).getTime();
    return time >= now && time <= threshold;
  }).length;
});

function selectActivity(id: string) {
  if (selectedActivityId.value === id) return;
  selectedActivityId.value = id;
}

const tokens = ref<ActivityTokenRecord[]>([]);
const tokensLoading = ref(false);
const viewerEntry = ref<ViewerEntryResponse | null>(null);
const viewerEntryLoading = ref(false);

async function loadManagementData(activityId: string) {
  const current = activityId;
  tokensLoading.value = true;
  viewerEntryLoading.value = true;
  try {
    const [tokenData, entryData] = await Promise.all([
      listActivityTokens(activityId),
      getViewerEntry(activityId)
    ]);
    if (selectedActivityId.value !== current) {
      return;
    }
    tokens.value = tokenData ?? [];
    viewerEntry.value = entryData ?? null;
  } catch (error) {
    if (error instanceof Error) {
      message?.error(error.message);
    } else {
      message?.error("加载管理数据失败，请稍后重试。");
    }
    if (selectedActivityId.value === current) {
      tokens.value = [];
      viewerEntry.value = null;
    }
  } finally {
    if (selectedActivityId.value === current) {
      tokensLoading.value = false;
      viewerEntryLoading.value = false;
    }
  }
}

watch(selectedActivityId, (id) => {
  if (!id) {
    tokens.value = [];
    viewerEntry.value = null;
    return;
  }
  loadManagementData(id);
});

const speakerTokens = computed(() => tokens.value.filter((item) => item.type === "speaker"));
const viewerTokens = computed(() => tokens.value.filter((item) => item.type === "viewer"));

function formatStatus(status: ActivityDto["status"]) {
  switch (status) {
    case "draft":
      return "草稿";
    case "published":
      return "已发布";
    case "closed":
      return "已关闭";
    default:
      return status;
  }
}

function formatLanguage(code: string) {
  return languageOptions.find((item) => item.code === code)?.label ?? code;
}

function formatTokenStatus(status: ActivityTokenRecord["status"]) {
  switch (status) {
    case "active":
      return "生效中";
    case "revoked":
      return "已撤销";
    case "expired":
      return "已过期";
    default:
      return status;
  }
}

function formatViewerStatus(status: ViewerEntryResponse["status"]) {
  switch (status) {
    case "active":
      return "已启用";
    case "inactive":
      return "未启用";
    case "revoked":
      return "已撤销";
    default:
      return status;
  }
}

function formatDateTime(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", {
    hour12: false,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function copyToClipboard(text: string, label: string) {
  if (!text) {
    message?.warning(`${label}为空`);
    return;
  }
  navigator.clipboard
    .writeText(text)
    .then(() => message?.success(`${label}已复制`))
    .catch(() => message?.error("复制失败，请检查浏览器权限"));
}

function nextDefaultStartTime() {
  const date = new Date();
  date.setMinutes(date.getMinutes() + 30);
  return toLocalInputValue(date);
}

function toLocalInputValue(value: string | Date) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const tzOffset = date.getTimezoneOffset();
  const local = new Date(date.getTime() - tzOffset * 60000);
  return local.toISOString().slice(0, 16);
}

function toISOString(value: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString();
}

const createForm = reactive({
  title: "",
  description: "",
  speaker: "",
  startTime: nextDefaultStartTime(),
  inputLanguage: "zh-CN",
  targetLanguages: ["en"] as string[],
  coverUrl: ""
});

function resetCreateForm() {
  createForm.title = "";
  createForm.description = "";
  createForm.speaker = "";
  createForm.startTime = nextDefaultStartTime();
  createForm.inputLanguage = "zh-CN";
  createForm.targetLanguages = ["en"];
  createForm.coverUrl = "";
}

const editForm = reactive({
  title: "",
  description: "",
  speaker: "",
  startTime: "",
  inputLanguage: "zh-CN",
  targetLanguages: [] as string[],
  coverUrl: ""
});

function populateEditForm(activity: ActivityDto) {
  editForm.title = activity.title;
  editForm.description = activity.description ?? "";
  editForm.speaker = activity.speaker;
  editForm.startTime = toLocalInputValue(activity.startTime);
  editForm.inputLanguage = activity.inputLanguage;
  editForm.targetLanguages = [...activity.targetLanguages];
  editForm.coverUrl = activity.coverUrl ?? "";
}

function toggleLanguage(bucket: string[], code: string) {
  const index = bucket.indexOf(code);
  if (index > -1) {
    bucket.splice(index, 1);
  } else {
    bucket.push(code);
  }
}

const showCreateModal = ref(false);
const showEditModal = ref(false);

function openCreateModal() {
  resetCreateForm();
  showCreateModal.value = true;
}

function closeCreateModal() {
  showCreateModal.value = false;
}

function openEditModal() {
  if (!selectedActivity.value) return;
  populateEditForm(selectedActivity.value);
  showEditModal.value = true;
}

function closeEditModal() {
  showEditModal.value = false;
}

const createMutation = useMutation({
  mutationFn: (payload: CreateActivityPayload) => createActivity(payload)
});

const updateMutation = useMutation({
  mutationFn: ({ id, payload }: { id: string; payload: UpdateActivityPayload }) =>
    updateActivity(id, payload)
});

const publishMutation = useMutation({
  mutationFn: (id: string) => publishActivity(id)
});

const closeMutation = useMutation({
  mutationFn: (id: string) => closeActivity(id)
});

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteActivity(id)
});

const speakerTokenMutation = useMutation({
  mutationFn: (id: string) => generateSpeakerToken(id)
});

const revokeSpeakerTokensMutation = useMutation({
  mutationFn: (id: string) => revokeSpeakerTokens(id)
});

const revokeSpeakerTokenMutation = useMutation({
  mutationFn: ({ id, tokenId }: { id: string; tokenId: string }) =>
    revokeSpeakerToken(id, tokenId)
});

const viewerTokenMutation = useMutation({
  mutationFn: ({ id, payload }: { id: string; payload: GenerateViewerTokenPayload }) =>
    generateViewerToken(id, payload)
});

const revokeEntryMutation = useMutation({
  mutationFn: (id: string) => revokeViewerEntry(id)
});

const activateEntryMutation = useMutation({
  mutationFn: (id: string) => activateViewerEntry(id)
});

const createPending = computed(() => createMutation.isPending.value);
const updatePending = computed(() => updateMutation.isPending.value);
const publishPending = computed(() => publishMutation.isPending.value);
const closePending = computed(() => closeMutation.isPending.value);
const deletePending = computed(() => deleteMutation.isPending.value);
const speakerTokenPending = computed(() => speakerTokenMutation.isPending.value);
const revokeSpeakerTokensPending = computed(() => revokeSpeakerTokensMutation.isPending.value);
const revokeSpeakerTokenPending = computed(() => revokeSpeakerTokenMutation.isPending.value);
const viewerTokenPending = computed(() => viewerTokenMutation.isPending.value);
const revokeEntryPending = computed(() => revokeEntryMutation.isPending.value);
const activateEntryPending = computed(() => activateEntryMutation.isPending.value);

async function submitCreate() {
  if (!createForm.title.trim() || !createForm.speaker.trim()) {
    message?.warning("请完整填写活动标题与主讲人。");
    return;
  }
  if (!createForm.startTime) {
    message?.warning("请选择开始时间。");
    return;
  }
  if (!createForm.targetLanguages.length) {
    message?.warning("至少选择一个目标语种。");
    return;
  }

  const payload: CreateActivityPayload = {
    title: createForm.title.trim(),
    description: createForm.description.trim(),
    speaker: createForm.speaker.trim(),
    startTime: toISOString(createForm.startTime),
    inputLanguage: createForm.inputLanguage,
    targetLanguages: [...createForm.targetLanguages],
    coverUrl: createForm.coverUrl.trim() || undefined
  };

  try {
    const created = await createMutation.mutateAsync(payload);
    message?.success("活动创建成功");
    showCreateModal.value = false;
    await refetchActivities();
    selectedActivityId.value = created.id;
  } catch (error) {
    const msg = error instanceof Error ? error.message : "创建失败，请稍后重试。";
    message?.error(msg);
  }
}

async function submitEdit() {
  if (!selectedActivity.value) return;
  if (!editForm.title.trim() || !editForm.speaker.trim()) {
    message?.warning("请完整填写活动标题与主讲人。");
    return;
  }
  if (!editForm.startTime) {
    message?.warning("请选择开始时间。");
    return;
  }
  if (!editForm.targetLanguages.length) {
    message?.warning("至少选择一个目标语种。");
    return;
  }

  const payload: UpdateActivityPayload = {
    title: editForm.title.trim(),
    description: editForm.description.trim(),
    speaker: editForm.speaker.trim(),
    startTime: toISOString(editForm.startTime),
    inputLanguage: editForm.inputLanguage,
    targetLanguages: [...editForm.targetLanguages],
    coverUrl: editForm.coverUrl.trim() || undefined
  };

  try {
    await updateMutation.mutateAsync({ id: selectedActivity.value.id, payload });
    message?.success("活动信息已更新");
    showEditModal.value = false;
    await refetchActivities();
    if (selectedActivityId.value) {
      await loadManagementData(selectedActivityId.value);
    }
  } catch (error) {
    const msg = error instanceof Error ? error.message : "保存失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handlePublish() {
  if (!selectedActivity.value) return;
  try {
    await publishMutation.mutateAsync(selectedActivity.value.id);
    message?.success("活动已发布");
    await refetchActivities();
    if (selectedActivityId.value) {
      await loadManagementData(selectedActivityId.value);
    }
  } catch (error) {
    const msg = error instanceof Error ? error.message : "发布失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleClose() {
  if (!selectedActivity.value) return;
  try {
    await closeMutation.mutateAsync(selectedActivity.value.id);
    message?.success("活动已关闭");
    await refetchActivities();
    if (selectedActivityId.value) {
      await loadManagementData(selectedActivityId.value);
    }
  } catch (error) {
    const msg = error instanceof Error ? error.message : "关闭失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleDelete() {
  if (!selectedActivity.value) return;
  if (!window.confirm("确认删除该草稿？该操作不可恢复。")) {
    return;
  }
  try {
    await deleteMutation.mutateAsync(selectedActivity.value.id);
    message?.success("活动已删除");
    selectedActivityId.value = null;
    await refetchActivities();
  } catch (error) {
    const msg = error instanceof Error ? error.message : "删除失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleGenerateSpeakerToken() {
  if (!selectedActivityId.value) return;
  try {
    const data = await speakerTokenMutation.mutateAsync(selectedActivityId.value);
    message?.success(`演讲者令牌已生成：${data.token}。请复制并发送给演讲者。`);
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "生成失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleRevokeSpeakerTokens() {
  if (!selectedActivityId.value) return;
  if (!speakerTokens.value.length) return;
  if (!window.confirm("确认撤销该活动下的所有演讲者令牌？撤销后需重新生成。")) {
    return;
  }
  try {
    await revokeSpeakerTokensMutation.mutateAsync(selectedActivityId.value);
    message?.success("已撤销演讲者令牌");
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "撤销失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleRevokeSingleSpeakerToken(token: ActivityTokenRecord) {
  if (!selectedActivityId.value) return;
  if (token.status !== "active") {
    message?.warning("该令牌已失效，无需重复撤销。");
    return;
  }
  if (!window.confirm("确认撤销该演讲者令牌？撤销后需重新生成新的令牌。")) {
    return;
  }
  try {
    await revokeSpeakerTokenMutation.mutateAsync({ id: selectedActivityId.value, tokenId: token.id });
    message?.success("演讲者令牌已撤销");
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "撤销失败，请稍后重试。";
    message?.error(msg);
  }
}

const viewerTokenForm = reactive({
  ttlMinutes: 120,
  maxAudience: undefined as number | undefined
});

async function handleGenerateViewerToken() {
  if (!selectedActivityId.value) return;
  const payload: GenerateViewerTokenPayload = {};
  if (viewerTokenForm.ttlMinutes && viewerTokenForm.ttlMinutes > 0) {
    payload.ttlMinutes = viewerTokenForm.ttlMinutes;
  }
  if (viewerTokenForm.maxAudience && viewerTokenForm.maxAudience > 0) {
    payload.maxAudience = viewerTokenForm.maxAudience;
  }
  try {
    const data = await viewerTokenMutation.mutateAsync({ id: selectedActivityId.value, payload });
    message?.success(`已生成观众邀请码：${data.code}`);
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "生成失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleRevokeEntry() {
  if (!selectedActivityId.value) return;
  try {
    await revokeEntryMutation.mutateAsync(selectedActivityId.value);
    message?.success("观众入口已撤销");
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "操作失败，请稍后重试。";
    message?.error(msg);
  }
}

async function handleActivateEntry() {
  if (!selectedActivityId.value) return;
  try {
    await activateEntryMutation.mutateAsync(selectedActivityId.value);
    message?.success("观众入口已启用");
    await loadManagementData(selectedActivityId.value);
  } catch (error) {
    const msg = error instanceof Error ? error.message : "操作失败，请稍后重试。";
    message?.error(msg);
  }
}
</script>

<style scoped>
.admin-console {
  display: flex;
  flex-direction: column;
  gap: 32px;
  padding: 8px 12px 48px;
}

.admin-console__hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28px 32px;
  border-radius: 28px;
  background: radial-gradient(circle at 10% 10%, rgba(16, 185, 129, 0.16), transparent),
    radial-gradient(circle at 85% 15%, rgba(14, 165, 233, 0.18), transparent),
    linear-gradient(135deg, rgba(15, 23, 42, 0.94), rgba(30, 64, 175, 0.85));
  color: #fff;
  box-shadow:
    0 34px 60px -40px rgba(15, 23, 42, 0.65),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.admin-console__hero h1 {
  margin: 0 0 6px;
  font-size: 26px;
  font-weight: 700;
}

.admin-console__hero p {
  margin: 0;
  max-width: 520px;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(226, 232, 240, 0.92);
}

.hero__action {
  border: none;
  background: linear-gradient(135deg, #34d399, #10b981);
  color: #0f172a;
  border-radius: 14px;
  padding: 12px 22px;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 20px 40px -32px rgba(16, 185, 129, 0.8);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.hero__action:hover {
  transform: translateY(-2px);
  box-shadow: 0 24px 48px -32px rgba(16, 185, 129, 0.9);
}

.admin-console__stats {
  display: grid;
  gap: 18px;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}

.stat-card {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 22px;
  padding: 20px 24px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  box-shadow:
    0 24px 48px -36px rgba(15, 23, 42, 0.45),
    0 1px 0 rgba(255, 255, 255, 0.6);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-card header {
  font-size: 14px;
  color: #475569;
}

.stat-card strong {
  font-size: 28px;
  color: #0f172a;
}

.stat-card span {
  font-size: 13px;
  color: #64748b;
}

.admin-console__grid {
  display: grid;
  gap: 28px;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
}

.admin-console__primary,
.admin-console__secondary {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.admin-panel {
  background: rgba(255, 255, 255, 0.94);
  border-radius: 26px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  padding: 26px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow:
    0 28px 52px -40px rgba(15, 23, 42, 0.5),
    0 1px 0 rgba(255, 255, 255, 0.7);
}

.admin-panel.empty {
  justify-content: center;
  align-items: center;
  min-height: 180px;
}

.admin-panel__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}

.admin-panel__header h2 {
  margin: 0 0 6px;
  font-size: 20px;
  color: #0f172a;
}

.admin-panel__header p {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.filter-group {
  display: flex;
  gap: 10px;
}

.filter-chip {
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(248, 250, 252, 0.9);
  border-radius: 12px;
  padding: 8px 14px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-chip.is-active {
  border-color: rgba(16, 185, 129, 0.8);
  background: rgba(209, 250, 229, 0.9);
  color: #047857;
  box-shadow: 0 16px 28px -24px rgba(16, 185, 129, 0.6);
}

.activity-list {
  display: grid;
  gap: 18px;
}

.activity-card {
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.32);
  padding: 22px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.94));
  cursor: pointer;
  transition: transform 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.24s cubic-bezier(0.4, 0, 0.2, 1),
    border 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}

.activity-card:hover {
  transform: translateY(-3px);
  border-color: rgba(16, 185, 129, 0.7);
  box-shadow:
    0 28px 46px -36px rgba(15, 23, 42, 0.55),
    0 18px 36px -34px rgba(20, 184, 166, 0.4);
}

.activity-card.is-active {
  border-color: rgba(16, 185, 129, 0.9);
  box-shadow:
    0 32px 52px -34px rgba(20, 184, 166, 0.55),
    0 1px 0 rgba(255, 255, 255, 0.75);
  background: linear-gradient(140deg, rgba(224, 255, 244, 0.8), rgba(255, 255, 255, 0.96));
}

.activity-card header {
  display: flex;
  justify-content: space-between;
  gap: 18px;
}

.activity-card h3 {
  margin: 0 0 4px;
  font-size: 18px;
  color: #0f172a;
}

.activity-meta {
  font-size: 13px;
  color: #64748b;
}

.badge {
  align-self: flex-start;
  border-radius: 12px;
  padding: 6px 12px;
  background: rgba(56, 189, 248, 0.18);
  color: #0f172a;
  font-size: 12px;
  font-weight: 600;
}

.activity-description {
  margin: 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
}

.lang-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.lang-chip {
  border-radius: 14px;
  padding: 6px 12px;
  background: rgba(226, 232, 240, 0.8);
  color: #1e293b;
  font-size: 12px;
  font-weight: 600;
}

.card-cta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #0f172a;
  font-weight: 600;
}

.card-cta svg {
  width: 16px;
  height: 16px;
}

.status-badge {
  border-radius: 999px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge[data-status="draft"] {
  background: rgba(253, 230, 138, 0.28);
  color: #92400e;
}

.status-badge[data-status="published"] {
  background: rgba(167, 243, 208, 0.34);
  color: #047857;
}

.status-badge[data-status="closed"] {
  background: rgba(248, 113, 113, 0.25);
  color: #b91c1c;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin: 0;
}

.detail-grid dt {
  font-size: 12px;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 4px;
}

.detail-grid dd {
  margin: 0;
  font-size: 14px;
  color: #0f172a;
  font-weight: 600;
}

.detail-section h3 {
  margin: 0 0 10px;
  font-size: 16px;
  color: #0f172a;
}

.detail-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.detail-actions button {
  border-radius: 12px;
  padding: 10px 16px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.detail-actions button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.detail-actions .primary {
  background: linear-gradient(135deg, #10b981, #0ea371);
  color: #fff;
  box-shadow: 0 16px 32px -24px rgba(16, 185, 129, 0.8);
}

.detail-actions .secondary {
  background: rgba(15, 23, 42, 0.05);
  color: #0f172a;
}

.detail-actions .danger {
  background: rgba(248, 113, 113, 0.18);
  color: #b91c1c;
}

.detail-actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  box-shadow: none;
}

.link-row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.link-row a {
  color: #0f172a;
  font-weight: 600;
  word-break: break-all;
}

.link-row .ghost {
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  background: rgba(248, 250, 252, 0.9);
  padding: 6px 12px;
  cursor: pointer;
}

.token-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.token-section__hint {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.token-section__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.token-section__header h3 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}

.token-section__actions {
  display: flex;
  gap: 10px;
}

.token-section__actions .secondary,
.token-section__actions .danger {
  border-radius: 12px;
  padding: 8px 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.token-section__actions .secondary {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.token-section__actions .danger {
  background: rgba(248, 113, 113, 0.18);
  color: #b91c1c;
}

.token-section__actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.token-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}

.token-list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: rgba(248, 250, 252, 0.9);
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.token-list strong {
  font-size: 15px;
  color: #0f172a;
}

.token-meta {
  display: block;
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
}

.token-list__actions {
  display: flex;
  gap: 8px;
}

.token-list .ghost {
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.36);
  background: rgba(255, 255, 255, 0.92);
  padding: 6px 12px;
  cursor: pointer;
  font-weight: 600;
}

.token-list__actions .danger {
  border-color: rgba(248, 113, 113, 0.45);
  color: #b91c1c;
  background: rgba(254, 226, 226, 0.9);
}

.token-list__actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.viewer-token-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  align-items: end;
}

.viewer-token-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 12px;
  color: #475569;
}

.viewer-token-form input {
  border: 1px solid rgba(148, 163, 184, 0.4);
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 14px;
  background: rgba(248, 250, 252, 0.9);
}

.viewer-token-form .secondary {
  border: none;
  border-radius: 12px;
  padding: 10px 16px;
  background: rgba(14, 165, 233, 0.18);
  color: #0369a1;
  font-weight: 600;
  cursor: pointer;
}

.viewer-token-form .secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.viewer-entry {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 14px;
  color: #0f172a;
}

.qr-preview {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.qr-preview textarea {
  width: 100%;
  min-height: 80px;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(248, 250, 252, 0.9);
  padding: 10px;
  font-family: "SFMono-Regular", "Menlo", monospace;
  font-size: 13px;
  color: #1e293b;
}

.qr-preview .ghost {
  align-self: flex-start;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.36);
  background: rgba(255, 255, 255, 0.92);
  padding: 6px 12px;
  cursor: pointer;
}

.viewer-entry__actions {
  display: flex;
  gap: 12px;
}

.viewer-entry__actions .secondary,
.viewer-entry__actions .danger {
  border: none;
  border-radius: 12px;
  padding: 10px 16px;
  font-weight: 600;
  cursor: pointer;
}

.viewer-entry__actions .secondary {
  background: rgba(96, 165, 250, 0.2);
  color: #1d4ed8;
}

.viewer-entry__actions .danger {
  background: rgba(248, 113, 113, 0.2);
  color: #b91c1c;
}

.viewer-entry__actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.muted {
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.6;
}

.modal {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
}

.modal__content {
  position: relative;
  width: min(620px, 92vw);
  max-height: 90vh;
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.98);
  border-radius: 24px;
  padding: 28px 30px;
  box-shadow:
    0 34px 68px -40px rgba(15, 23, 42, 0.65),
    0 1px 0 rgba(255, 255, 255, 0.8);
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.modal__content header h2 {
  margin: 0 0 6px;
  font-size: 22px;
  color: #0f172a;
}

.modal__content header p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-form label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
  color: #1e293b;
}

.modal-form input[type="datetime-local"],
.modal-form select {
  border: 1px solid rgba(148, 163, 184, 0.36);
  border-radius: 12px;
  padding: 10px 12px;
  font-size: 14px;
  background: rgba(248, 250, 252, 0.92);
}

.language-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.language-chip {
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  padding: 6px 12px;
  font-size: 13px;
  background: rgba(248, 250, 252, 0.9);
  cursor: pointer;
  transition: all 0.2s ease;
}

.language-chip.is-selected {
  border-color: rgba(16, 185, 129, 0.8);
  background: rgba(209, 250, 229, 0.9);
  color: #047857;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
}

.modal-actions button {
  border-radius: 12px;
  padding: 10px 18px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}

.modal-actions .ghost {
  background: rgba(248, 250, 252, 0.92);
  border: 1px solid rgba(148, 163, 184, 0.36);
  color: #1e293b;
}

.modal-actions .primary {
  background: linear-gradient(135deg, #10b981, #0ea371);
  color: #fff;
}

.modal-actions .primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 1280px) {
  .admin-console__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .admin-console__hero {
    flex-direction: column;
    align-items: flex-start;
    gap: 18px;
  }

  .hero__action {
    width: 100%;
    text-align: center;
  }

  .admin-panel__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .detail-actions,
  .viewer-entry__actions {
    flex-direction: column;
    width: 100%;
  }

  .detail-actions button,
  .viewer-entry__actions button {
    width: 100%;
    text-align: center;
  }

  .viewer-token-form {
    grid-template-columns: 1fr;
  }
}
</style>
