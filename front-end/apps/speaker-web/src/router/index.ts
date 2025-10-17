import { createRouter, createWebHistory } from "vue-router";
import SpeakerConsole from "../views/SpeakerConsole.vue";
import LoginView from "../views/LoginView.vue";
import AdminDashboard from "../views/AdminDashboard.vue";
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
    },
    {
      path: "/admin",
      name: "admin-dashboard",
      component: AdminDashboard,
      meta: {
        requiresAuth: true,
        requiresAdmin: true,
        shell: false
      }
    }
  ]
});

function isAdminUser(username?: string | null) {
  return (username ?? "").toLowerCase() === "admin";
}

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore();
  if (to.meta.requiresAuth === false) {
    if (auth.isAuthenticated && to.name === "login") {
      const target = isAdminUser(auth.profile?.username) ? "admin-dashboard" : "speaker-console";
      next({ name: target });
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

  if (to.meta.requiresAdmin) {
    if (!isAdminUser(auth.profile?.username)) {
      next({ name: "speaker-console" });
      return;
    }
  }

  next();
});

export default router;
