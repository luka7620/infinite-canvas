import type { CSSProperties } from "react";
import type { ThemeConfig } from "antd";
import { theme as antdTheme } from "antd";

const palette = {
    light: {
        primary: "#0f6f68",
        primaryHover: "#0a5752",
        primaryText: "#ffffff",
        bgLayout: "#f5f6f5",
        bgContainer: "#ffffff",
        border: "rgba(44, 54, 60, 0.16)",
        text: "#1f2428",
        textSecondary: "#616a70",
        menuBg: "rgba(15, 111, 104, 0.1)",
        menuText: "#0f6f68",
        selectActiveBg: "rgba(15, 111, 104, 0.08)",
        selectSelectedBg: "rgba(15, 111, 104, 0.13)",
        selectText: "#1f2428",
        tableSelectedBg: "rgba(15, 111, 104, 0.08)",
        tableSelectedHoverBg: "rgba(15, 111, 104, 0.12)",
    },
    dark: {
        primary: "#69d8c8",
        primaryHover: "#9beade",
        primaryText: "#111817",
        bgLayout: "#121414",
        bgContainer: "#1b1e1d",
        border: "rgba(178, 187, 184, 0.17)",
        text: "#f0f3f2",
        textSecondary: "#b0b8b5",
        menuBg: "rgba(105, 216, 200, 0.12)",
        menuText: "#c8fff4",
        selectActiveBg: "rgba(105, 216, 200, 0.1)",
        selectSelectedBg: "rgba(105, 216, 200, 0.16)",
        selectText: "#f0f3f2",
        tableSelectedBg: "rgba(105, 216, 200, 0.1)",
        tableSelectedHoverBg: "rgba(105, 216, 200, 0.16)",
    },
};

export const adminLayoutStyle = {
    siderWidth: 232,
    headerHeight: 56,
    brandHeight: 64,
    menu: { borderInlineEnd: 0, padding: "18px 12px", fontSize: 15 } satisfies CSSProperties,
    menuItem: { height: 44, lineHeight: "44px", marginBlock: 4, borderRadius: 8 } satisfies CSSProperties,
};

export function getAntThemeConfig(dark: boolean): ThemeConfig {
    const color = dark ? palette.dark : palette.light;

    return {
        algorithm: dark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        cssVar: { key: dark ? "infinite-canvas-dark" : "infinite-canvas-light" },
        token: {
            borderRadius: 8,
            borderRadiusLG: 8,
            colorBgContainer: color.bgContainer,
            colorBgLayout: color.bgLayout,
            colorBorder: color.border,
            colorPrimary: color.primary,
            colorInfo: color.primary,
            colorLink: color.primary,
            colorLinkHover: color.primaryHover,
            colorLinkActive: color.primary,
            colorText: color.text,
            colorTextBase: color.text,
            colorTextLightSolid: color.primaryText,
            colorTextSecondary: color.textSecondary,
            fontFamily: '"SF Pro Display","SF Pro Text","PingFang SC","Microsoft YaHei","Helvetica Neue",sans-serif',
        },
        components: {
            Button: {
                primaryShadow: "none",
            },
            Menu: {
                itemActiveBg: color.menuBg,
                itemHoverBg: color.menuBg,
                itemSelectedBg: color.menuBg,
                itemSelectedColor: color.menuText,
                darkItemHoverBg: palette.dark.menuBg,
                darkItemSelectedBg: palette.dark.menuBg,
                darkItemSelectedColor: palette.dark.menuText,
            },
            Select: {
                optionActiveBg: color.selectActiveBg,
                optionSelectedBg: color.selectSelectedBg,
                optionSelectedColor: color.selectText,
            },
            Table: {
                rowSelectedBg: color.tableSelectedBg,
                rowSelectedHoverBg: color.tableSelectedHoverBg,
            },
        },
    };
}
