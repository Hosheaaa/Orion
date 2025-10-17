import { createRouter, createWebHistory } from "vue-router";
import SpeakerConsole from "../views/SpeakerConsole.vue";
import LoginView from "../views/LoginView.vue";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: LoginView,
      meta: {
        requiresAuth: false
      }
    },
    {
      path: "/",
      name: "speaker-console",
      component: SpeakerConsole,
      meta: {
        requiresAuth: true
      }
    }
  ]
});

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore();
  if (to.meta.requiresAuth === false) {
    if (auth.isAuthenticated && to.name === "login") {
      next({ name: "speaker-console" });
      return;
    }
    next();
    return;
  }

  if (!auth.isAuthenticated) {
    next({ name: "login" });
    return;
  }

  if (auth.shouldRefresh()) {
    try {
      await auth.refresh();
    } catch {
      auth.logout();
      next({ name: "login" });
      return;
    }
  }

  next();
});

export default router;
