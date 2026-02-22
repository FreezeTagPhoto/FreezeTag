"use client";

// When adding a theme, add it to this registry and also add it to themes.css
export const DarkThemeRegistry = [
    "Catppuccin Frappe",
    "Catppuccin Macchiato",
    "Catppuccin Mocha",
];
export const LightThemeRegistry = [
    "Catppuccin Latte",
    "Microsoft Hot Dog Stand",
];

export const THEME_STORAGE_KEY = "freezetag-theme-option";

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
    if (DarkThemeRegistry.includes(theme) || LightThemeRegistry.includes(theme)) {
        localStorage.setItem(THEME_STORAGE_KEY, theme);
    } else {
        console.error("Attempted to set invalid theme:", theme);
    }
};

export const ApplyTheme = (theme: string) => {
    if (typeof document === "undefined") return;

    document.documentElement.setAttribute("data-theme", theme);

    const type = DarkThemeRegistry.includes(theme) ? "dark" : "light";
    document.documentElement.setAttribute("data-theme-type", type);
};
