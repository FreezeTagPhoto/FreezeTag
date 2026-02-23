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
            <header className={styles.header}>
                <h1 className={styles.h1}>Settings</h1>
                <p className={styles.subtitle}>
                    Customize preferences and manage account security.
                </p>

                {/* TODO: enable this when we have more settings options. */}
                {/* <nav className={styles.jumpNav} aria-label="Settings sections">
                    <a className={styles.jumpLink} href="#preferences">
                        Preferences
                    </a>
                    <a className={styles.jumpLink} href="#security">
                        Security
                    </a>
                </nav> */}

            </header>

            <section
                id="preferences"
                className={styles.panel}
                aria-labelledby="preferences-heading"
            >
                <div className={styles.sectionHeader}>
                    <h2
                        id="preferences-heading"
                        className={styles.sectionTitle}
                    >
                        Preferences
                    </h2>
                    <p className={styles.sectionDescription}>
                        Appearance and measurement defaults.
                    </p>
                </div>

                <div className={styles.fields}>
                    <div className={styles.fieldRow}>
                        <div className={styles.fieldText}>
                            <label className={styles.label} htmlFor="theme">
                                Theme
                            </label>
                            <p className={styles.hint}>
                                Choose a light or dark theme.
                            </p>
                        </div>

                        <div className={styles.control}>
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
                    </div>

                    <div className={styles.fieldRow}>
                        <div className={styles.fieldText}>
                            <div id="units-label" className={styles.label}>
                                Units
                            </div>
                            <p className={styles.hint}>
                                Controls how distances are displayed.
                            </p>
                        </div>

                        <div className={styles.control}>
                            <div
                                className={styles.segmented}
                                role="radiogroup"
                                aria-labelledby="units-label"
                            >
                                <label className={styles.segment}>
                                    <input
                                        type="radio"
                                        name="units"
                                        value="metric"
                                        checked={units === "metric"}
                                        onChange={() => {
                                            const next: UnitSystem = "metric";
                                            setUnits(next);
                                            UnitsSetter(next);
                                            ApplyUnits(next);
                                            window.dispatchEvent(
                                                new Event(
                                                    "freezetag:units-changed",
                                                ),
                                            );
                                        }}
                                    />
                                    <span className={styles.segmentText}>
                                        Metric
                                    </span>
                                </label>

                                <label className={styles.segment}>
                                    <input
                                        type="radio"
                                        name="units"
                                        value="imperial"
                                        checked={units === "imperial"}
                                        onChange={() => {
                                            const next: UnitSystem = "imperial";
                                            setUnits(next);
                                            UnitsSetter(next);
                                            ApplyUnits(next);
                                            window.dispatchEvent(
                                                new Event(
                                                    "freezetag:units-changed",
                                                ),
                                            );
                                        }}
                                    />
                                    <span className={styles.segmentText}>
                                        Imperial
                                    </span>
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            <section
                id="security"
                className={styles.panel}
                aria-labelledby="security-heading"
            >
                <div className={styles.sectionHeader}>
                    <h2 id="security-heading" className={styles.sectionTitle}>
                        Security
                    </h2>
                    <p className={styles.sectionDescription}>
                        Sign-in and password settings.
                    </p>
                </div>

                <div className={styles.fields}>
                    <div className={styles.fieldRow}>
                        <div className={styles.fieldText}>
                            <div className={styles.label}>Password</div>
                            <p className={styles.hint}>
                                Change your account password.
                            </p>
                        </div>

                        <div className={styles.control}>
                            <button
                                type="button"
                                className={styles.button}
                                disabled
                            >
                                Coming soon
                            </button>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}
