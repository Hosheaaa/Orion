import { createRouter, createWebHistory } from "vue-router";
import SpeakerConsole from "../views/SpeakerConsole.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "speaker-console",
      component: SpeakerConsole
    }
  ]
});

export default router;
