"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./page.module.css";
import {
    DarkThemeRegistry,
    LightThemeRegistry,
    ThemeGetter,
    ThemeSetter,
    ApplyTheme,
} from "@/themes/ThemeManager";
import {
    ApplyUnits,
    UnitsGetter,
    UnitsSetter,
    type UnitSystem,
} from "@/common/units/UnitManager";

const ALL_THEMES = [...LightThemeRegistry, ...DarkThemeRegistry];
type ThemeName = (typeof ALL_THEMES)[number];

export default function SettingsPage() {
    const fallbackTheme = useMemo<ThemeName>(() => "Catppuccin Mocha", []);
    const [theme, setTheme] = useState<ThemeName>(fallbackTheme);

    const fallbackUnits = useMemo<UnitSystem>(() => "metric", []);
    const [units, setUnits] = useState<UnitSystem>(fallbackUnits);

    useEffect(() => {
        const initialTheme = ThemeGetter() as ThemeName;
        setTheme(initialTheme);
        ApplyTheme(initialTheme);

        const initialUnits = UnitsGetter();
        setUnits(initialUnits);
        ApplyUnits(initialUnits);
    }, [fallbackTheme, fallbackUnits]);

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
                            ThemeSetter(next);
                            ApplyTheme(next);
                            window.dispatchEvent(
                                new Event("freezetag:theme-changed"),
                            );
                        }}
                    >
                        <optgroup label="Light Themes">
                            {[...LightThemeRegistry].sort().map((t) => (
                                <option key={t} value={t}>
                                    {t}
                                </option>
                            ))}
                        </optgroup>

                        <optgroup label="Dark Themes">
                            {[...DarkThemeRegistry].sort().map((t) => (
                                <option key={t} value={t}>
                                    {t}
                                </option>
                            ))}
                        </optgroup>
                    </select>
                </div>

                <div className={styles.row}>
                    <label className={styles.label}>Units</label>

                    <select
                        id="units"
                        className={styles.select}
                        value={units}
                        onChange={(e) => {
                            const next = e.target.value as UnitSystem;

                            setUnits(next);
                            UnitsSetter(next);
                            ApplyUnits(next);
                            window.dispatchEvent(
                                new Event("freezetag:units-changed"),
                            );
                        }}
                    >
                        <option value="metric">Metric (km, m)</option>
                        <option value="imperial">Imperial (mi, ft, yd)</option>
                    </select>
                </div>
            </section>
        </main>
    );
}
