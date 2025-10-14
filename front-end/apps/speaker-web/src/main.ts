import { createApp } from "vue";
import { createPinia } from "pinia";
import { VueQueryPlugin } from "@tanstack/vue-query";
import App from "./App.vue";
import router from "./router";
import { naiveDiscrete } from "./plugins/naive";
import { queryClient } from "./query/client";
import "./styles/main.css";

const app = createApp(App);

const pinia = createPinia();
app.use(pinia);

app.use(router);
app.use(VueQueryPlugin, {
  queryClient
});

naiveDiscrete.install(app);

app.mount("#app");
