"use client";
// When adding a theme, add it to this registry and also add it to themes.css
export const DarkThemeRegistry = [
    "Catppuccin Frappe",
    "Catppuccin Macchiato",
    "Catppuccin Mocha",
];
export const LightThemeRegistry = [
    "Catppuccin Latte",
    "Microsoft Hotdog Stand",
];
export const THEME_STORAGE_KEY = "freezetag-theme-option";

// Returns the string for the proper selected theme. Should be put into a `data-theme` element in the most base element of a page
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

// Returns the string for the proper selected theme type. Should be put into a `data-theme-type` element in the most base element of a page
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
