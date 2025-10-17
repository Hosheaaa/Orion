<template>
  <div class="auth-page">
    <section class="auth-card">
      <header>
        <h1>Orion 实时演讲控制台</h1>
        <p>请使用后台分配的管理员或演讲者账号登录，以访问活动日程与推流能力。</p>
      </header>
      <form @submit.prevent="handleSubmit">
        <label>
          <span>用户名</span>
          <n-input v-model:value="form.username" placeholder="请输入用户名" />
        </label>
        <label>
          <span>密码</span>
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入密码"
          />
        </label>
        <button
          class="auth-submit"
          type="submit"
          :disabled="auth.loading"
        >
          {{ auth.loading ? "登录中..." : "登录" }}
        </button>
        <p v-if="error" class="auth-error">{{ error }}</p>
      </form>
    </section>
  </div>
</template>

<script setup lang="ts">
import { inject, reactive, ref } from "vue";
import type { MessageApiInjection } from "naive-ui/es/message/src/MessageProvider";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();
const message = inject<MessageApiInjection | undefined>("naive-message");

const form = reactive({
  username: "",
  password: ""
});

const error = ref("");

async function handleSubmit() {
  error.value = "";
  if (!form.username || !form.password) {
    error.value = "请填写完整的用户名与密码。";
    return;
  }
  try {
    await auth.login({
      username: form.username,
      password: form.password
    });
    message?.success("登录成功，正在跳转...");
    router.replace({ name: "speaker-console" });
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    error.value = msg || "登录失败，请稍后重试。";
    message?.error(error.value);
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 20% 20%, rgba(16, 185, 129, 0.15), transparent),
    radial-gradient(circle at 80% 10%, rgba(14, 165, 233, 0.15), transparent),
    #f8fafc;
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: white;
  padding: 32px 36px;
  border-radius: 24px;
  box-shadow:
    0 24px 48px -32px rgba(15, 23, 42, 0.2),
    0 1px 0 rgba(255, 255, 255, 0.6);
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.auth-card header h1 {
  margin: 0;
  font-size: 22px;
  color: #0f172a;
}

.auth-card header p {
  margin: 8px 0 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
}

form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
  color: #1e293b;
}

.auth-submit {
  margin-top: 6px;
  width: 100%;
  border: none;
  background: linear-gradient(135deg, #10b981, #0ea371);
  color: white;
  border-radius: 12px;
  padding: 12px 0;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.auth-submit:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.auth-error {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}
</style>
