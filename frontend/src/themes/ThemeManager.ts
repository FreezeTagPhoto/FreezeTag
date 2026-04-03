"use client";

// When adding a theme, add it to this registry and also add it to themes.css
export const DarkThemeRegistry = [
    "Catppuccin Frappe",
    "Catppuccin Macchiato",
    "Catppuccin Mocha",
    "Colorblind Mocha",
    "Custom Dark",
    "High Contrast Dark",
];
export const LightThemeRegistry = [
    "Catppuccin Latte",
    "Colorblind Latte",
    "Custom Light",
    "High Contrast Light",
    "Microsoft Hot Dog Stand",
];

export const THEME_STORAGE_KEY = "freezetag-theme-option";
export const CUSTOM_DARK_COLORS_KEY = "freezetag-custom-dark-colors";
export const CUSTOM_LIGHT_COLORS_KEY = "freezetag-custom-light-colors";

// catppuccin palette accent variables
export const ACCENT_VARIABLES = [
    "rosewater",
    "flamingo",
    "pink",
    "mauve",
    "red",
    "maroon",
    "peach",
    "yellow",
    "green",
    "teal",
    "sky",
    "sapphire",
    "blue",
    "lavender",
] as const;

// ui accent variables
export const UI_ACCENT_VARIABLES = ["accent1", "accent2", "accent3"] as const;

export type AccentVariable = (typeof ACCENT_VARIABLES)[number];
export type UIAccentVariable = (typeof UI_ACCENT_VARIABLES)[number];
export type AllCustomVariable = AccentVariable | UIAccentVariable;
export type CustomColors = Record<AllCustomVariable, string>;

export const MOCHA_ACCENT_DEFAULTS: CustomColors = {
    // catppuccin mocha
    rosewater: "#f5e0dc",
    flamingo: "#f2cdcd",
    pink: "#f5c2e7",
    mauve: "#cba6f7",
    red: "#f38ba8",
    maroon: "#eba0ac",
    peach: "#fab387",
    yellow: "#f9e2af",
    green: "#a6e3a1",
    teal: "#94e2d5",
    sky: "#89dceb",
    sapphire: "#74c7ec",
    blue: "#89b4fa",
    lavender: "#b4befe",

    // dark theme ui accents
    accent1: "#579dd7",
    accent2: "#aedbf0",
    accent3: "#ffffff",
};

export const LATTE_ACCENT_DEFAULTS: CustomColors = {
    // catppuccin latte
    rosewater: "#dc8a78",
    flamingo: "#dd7878",
    pink: "#ea76cb",
    mauve: "#8839ef",
    red: "#d20f39",
    maroon: "#e64553",
    peach: "#fe640b",
    yellow: "#df8e1d",
    green: "#40a02b",
    teal: "#179299",
    sky: "#04a5e5",
    sapphire: "#209fb5",
    blue: "#1e66f5",
    lavender: "#7287fd",

    // light theme ui accents
    accent1: "#aedbf0",
    accent2: "#579dd7",
    accent3: "#ffffff",
};

// neutral colors (hidden from users)
const MOCHA_NEUTRALS: Record<string, string> = {
    text: "#cdd6f4",
    subtext1: "#bac2de",
    subtext0: "#a6adc8",
    overlay2: "#9399b2",
    overlay1: "#7f849c",
    overlay0: "#6c7086",
    surface2: "#585b70",
    surface1: "#45475a",
    surface0: "#313244",
    base: "#1e1e2e",
    mantle: "#181825",
    crust: "#11111b",
};

const LATTE_NEUTRALS: Record<string, string> = {
    text: "#4c4f69",
    subtext1: "#5c5f77",
    subtext0: "#6c6f85",
    overlay2: "#7c7f93",
    overlay1: "#8c8fa1",
    overlay0: "#9ca0b0",
    surface2: "#acb0be",
    surface1: "#bcc0cc",
    surface0: "#ccd0da",
    base: "#eff1f5",
    mantle: "#e6e9ef",
    crust: "#dce0e8",
};

export const ThemeGetter: () => string = () => {
    const stored_theme = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored_theme) {
        if (
            DarkThemeRegistry.includes(stored_theme) ||
            LightThemeRegistry.includes(stored_theme)
        ) {
            return stored_theme;
        } else {
            console.error("Detected incorrect stored theme!");
        }
    }

    const default_light = window.matchMedia(
        "(prefers-color-scheme: light)",
    ).matches;
    return default_light ? "Catppuccin Latte" : "Catppuccin Mocha";
};

export const ThemeTypeGetter: () => string = () => {
    const stored_theme = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored_theme) {
        if (DarkThemeRegistry.includes(stored_theme)) {
            return "dark";
        } else if (LightThemeRegistry.includes(stored_theme)) {
            return "light";
        } else {
            console.error("Detected incorrect stored theme!");
        }
    }

    const default_light = window.matchMedia(
        "(prefers-color-scheme: light)",
    ).matches;
    return default_light ? "light" : "dark";
};

export const ThemeSetter = (theme: string) => {
    if (
        DarkThemeRegistry.includes(theme) ||
        LightThemeRegistry.includes(theme)
    ) {
        localStorage.setItem(THEME_STORAGE_KEY, theme);
    } else {
        console.error("Attempted to set invalid theme:", theme);
    }
};

export const CustomColorsGetter = (type: "dark" | "light"): CustomColors => {
    const storageKey =
        type === "dark" ? CUSTOM_DARK_COLORS_KEY : CUSTOM_LIGHT_COLORS_KEY;
    const stored = localStorage.getItem(storageKey);
    const defaults =
        type === "dark" ? MOCHA_ACCENT_DEFAULTS : LATTE_ACCENT_DEFAULTS;
    if (stored) {
        try {
            const parsed = JSON.parse(stored);
            const merged = { ...defaults };
            for (const v of [...ACCENT_VARIABLES, ...UI_ACCENT_VARIABLES]) {
                if (typeof parsed[v] === "string") {
                    merged[v] = parsed[v];
                }
            }
            return merged;
        } catch {
            // fall through to defaults
        }
    }
    return { ...defaults };
};

export const CustomColorsSetter = (
    type: "dark" | "light",
    colors: CustomColors,
) => {
    const storageKey =
        type === "dark" ? CUSTOM_DARK_COLORS_KEY : CUSTOM_LIGHT_COLORS_KEY;
    localStorage.setItem(storageKey, JSON.stringify(colors));
};

// applies given colors by injecting <style> tag with css variables for the given theme
export const ApplyCustomThemeColors = (
    themeName: "Custom Dark" | "Custom Light",
    colors: CustomColors,
) => {
    if (typeof document === "undefined") return;
    const neutrals =
        themeName === "Custom Dark" ? MOCHA_NEUTRALS : LATTE_NEUTRALS;
    const allVars = { ...neutrals, ...colors };
    const varBlock = Object.entries(allVars)
        .map(([k, v]) => `  --${k}: ${v};`)
        .join("\n");
    let styleEl = document.getElementById(
        "freezetag-custom-theme",
    ) as HTMLStyleElement | null;
    if (!styleEl) {
        styleEl = document.createElement("style");
        styleEl.id = "freezetag-custom-theme";
        document.head.appendChild(styleEl);
    }
    styleEl.textContent = `[data-theme="${themeName}"] {\n${varBlock}\n}`;
};

export const ApplyTheme = (theme: string) => {
    if (typeof document === "undefined") return;

    document.documentElement.setAttribute("data-theme", theme);

    const type = DarkThemeRegistry.includes(theme) ? "dark" : "light";
    document.documentElement.setAttribute("data-theme-type", type);

    if (theme === "Custom Dark" || theme === "Custom Light") {
        const colorType = theme === "Custom Dark" ? "dark" : "light";
        const colors = CustomColorsGetter(colorType);
        ApplyCustomThemeColors(theme, colors);
    }
};
