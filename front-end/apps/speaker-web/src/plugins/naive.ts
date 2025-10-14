import { App } from "vue";
import {
  create,
  createDiscreteApi,
  lightTheme,
  GlobalThemeOverrides,
  NButton,
  NCard,
  NConfigProvider,
  NDivider,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NFlex,
  NIcon,
  NInput,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NList,
  NListItem,
  NProgress,
  NSkeleton,
  NStatistic,
  NSwitch,
  NTag,
  NText,
  NTimeline,
  NTimelineItem,
  NTooltip
} from "naive-ui";

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: "#10b981",
    primaryColorHover: "#0ea371",
    primaryColorPressed: "#0b8d63",
    primaryColorSuppl: "#34d399",
    borderRadius: "12px",
    fontFamily:
      '"Inter", "PingFang SC", "Microsoft YaHei", "Helvetica Neue", Arial, sans-serif'
  },
  Card: {
    paddingMedium: "20px 24px"
  },
  Button: {
    borderRadius: "10px",
    heightMedium: "40px",
    fontWeight: "600"
  }
};

export const naiveDiscrete = {
  install(app: App) {
    const naive = create({
      components: [
        NButton,
        NCard,
        NConfigProvider,
        NDivider,
        NDrawer,
        NDrawerContent,
        NDropdown,
        NFlex,
        NIcon,
        NInput,
        NLayout,
        NLayoutContent,
        NLayoutHeader,
        NLayoutSider,
        NList,
        NListItem,
        NProgress,
        NSkeleton,
        NStatistic,
        NSwitch,
        NTag,
        NText,
        NTimeline,
        NTimelineItem,
        NTooltip
      ]
    });

    app.use(naive);

    const configProviderProps = {
      theme: lightTheme,
      themeOverrides
    };

    app.provide("naive-theme-config", configProviderProps);

    const { message, notification, dialog } = createDiscreteApi(
      ["message", "notification", "dialog"],
      {
        configProviderProps,
        theme: lightTheme,
        themeOverrides
      }
    );

    app.provide("naive-message", message);
    app.provide("naive-notification", notification);
    app.provide("naive-dialog", dialog);
  }
};
