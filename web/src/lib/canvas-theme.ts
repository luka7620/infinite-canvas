export type CanvasColorTheme = "light" | "dark";
export type CanvasBackgroundMode = "dots" | "lines" | "blank";

export const canvasThemes = {
    light: {
        canvas: {
            background: "#f5f6f5",
            dot: "rgba(44,54,60,.24)",
            line: "rgba(44,54,60,.11)",
            selectionStroke: "#0f6f68",
            selectionFill: "rgba(15,111,104,.08)",
        },
        node: {
            label: "#616a70",
            fill: "#eef1ef",
            panel: "#ffffff",
            stroke: "#d4ddda",
            activeStroke: "#0f6f68",
            placeholder: "#828d8a",
            text: "#1f2428",
            muted: "#616a70",
            faint: "#9ba4a1",
            danger: "#b4232a",
            asset: "#0f6f68",
            config: "#6b7d22",
        },
        toolbar: {
            panel: "rgba(255,255,255,.96)",
            border: "#d4ddda",
            item: "#616a70",
            itemHover: "#eef1ef",
            activeBg: "#dcebe8",
            activeText: "#0f3f3b",
        },
    },
    dark: {
        canvas: {
            background: "#121414",
            dot: "rgba(240,243,242,.2)",
            line: "rgba(240,243,242,.09)",
            selectionStroke: "#69d8c8",
            selectionFill: "rgba(105,216,200,.10)",
        },
        node: {
            label: "#b0b8b5",
            fill: "#242827",
            panel: "#1b1e1d",
            stroke: "#3c4441",
            activeStroke: "#69d8c8",
            placeholder: "#828e8a",
            text: "#f0f3f2",
            muted: "#b0b8b5",
            faint: "#6f7c78",
            danger: "#ff8d8d",
            asset: "#69d8c8",
            config: "#d4e96b",
        },
        toolbar: {
            panel: "rgba(27,30,29,.96)",
            border: "#3c4441",
            item: "#b0b8b5",
            itemHover: "#242827",
            activeBg: "#263f3b",
            activeText: "#f0f3f2",
        },
    },
} as const;

export type CanvasTheme = (typeof canvasThemes)[CanvasColorTheme];
