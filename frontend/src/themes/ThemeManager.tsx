"use client";
// When adding a theme, add it to this registry and also add it to themes.css
export const ThemeRegistry = [
    "Catppuccin Mocha",
    "Catppuccin Macchiato",
    "Catppuccin Frappe",
    "Catppuccin Latte",
];

// Returns the string for the proper selected theme. Should be put into a `data-theme` element in the most base element of a page
export const ThemeGetter: () => string = () => {
    const stored_theme = localStorage.getItem("freezetag-theme-option");
    if (stored_theme) {
        if (ThemeRegistry.includes(stored_theme)) {
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
