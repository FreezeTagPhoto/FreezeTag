"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./page.module.css";
import {
    DarkThemeRegistry,
    LightThemeRegistry,
    ThemeGetter,
} from "@/themes/ThemeManager";
import { THEME_STORAGE_KEY as STORAGE_KEY } from "@/themes/ThemeManager";

const ALL_THEMES = [...LightThemeRegistry, ...DarkThemeRegistry];

type ThemeName = (typeof ALL_THEMES)[number];

function applyTheme(theme: ThemeName) {
    const root = document.documentElement;

    root.setAttribute("data-theme", theme);

    const type = DarkThemeRegistry.includes(theme) ? "dark" : "light";
    root.setAttribute("data-theme-type", type);
}

export default function SettingsPage() {
    const fallback = useMemo<ThemeName>(() => "Catppuccin Mocha", []);
    const [theme, setTheme] = useState<ThemeName>(fallback);

    useEffect(() => {
        const initial = ThemeGetter() as ThemeName;
        setTheme(initial);
        applyTheme(initial);
    }, [fallback]);

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Settings</h1>

            <section className={styles.section}>
                <div className={styles.row}>
                    <label className={styles.label}>Theme</label>

                    <select
                        id="theme"
                        className={styles.select}
                        value={theme}
                        onChange={(e) => {
                            const next = e.target.value as ThemeName;

                            setTheme(next);
                            localStorage.setItem(STORAGE_KEY, next);
                            applyTheme(next);
                        }}
                    >
                        {ALL_THEMES.map((t) => (
                            <option key={t} value={t}>
                                {t}
                            </option>
                        ))}
                    </select>
                </div>

                <p className={styles.hint}>
                    (this is just a proof of concept for now)
                </p>
            </section>
        </main>
    );
}
