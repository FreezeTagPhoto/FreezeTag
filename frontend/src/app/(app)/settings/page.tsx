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
import PasswordChanger from "@/api/auth/passwordchanger";
import { Option, Some, None } from "@/common/option";
import { Result, Ok, Err } from "@/common/result";
import { Eye, EyeOff } from "lucide-react";

const ALL_THEMES = [...LightThemeRegistry, ...DarkThemeRegistry];
type ThemeName = (typeof ALL_THEMES)[number];

export default function SettingsPage() {
    const fallbackTheme = useMemo<ThemeName>(() => "Catppuccin Mocha", []);
    const [theme, setTheme] = useState<ThemeName>(fallbackTheme);

    const fallbackUnits = useMemo<UnitSystem>(() => "metric", []);
    const [units, setUnits] = useState<UnitSystem>(fallbackUnits);

    const [currentPassword, setCurrentPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [pwBusy, setPwBusy] = useState(false);
    const [pwStatus, setPwStatus] =
        useState<Option<Result<string, string>>>(None());

    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    useEffect(() => {
        const initialTheme = ThemeGetter() as ThemeName;
        setTheme(initialTheme);
        ApplyTheme(initialTheme);

        const initialUnits = UnitsGetter();
        setUnits(initialUnits);
        ApplyUnits(initialUnits);
    }, [fallbackTheme, fallbackUnits]);

    const onChangePassword: React.FormEventHandler<HTMLFormElement> = async (
        e,
    ) => {
        e.preventDefault();
        setPwStatus(None());

        if (!currentPassword || !newPassword || !confirmPassword) {
            setPwStatus(Some(Err("Please fill out all fields.")));
            return;
        }

        if (newPassword !== confirmPassword) {
            setPwStatus(Some(Err("New passwords do not match.")));
            return;
        }

        if (newPassword === currentPassword) {
            setPwStatus(
                Some(
                    Err(
                        "New password must be different from your current password.",
                    ),
                ),
            );
            return;
        }

        setPwBusy(true);

        const fd = new FormData();
        fd.set("current_password", currentPassword);
        fd.set("new_password", newPassword);

        const res = await PasswordChanger(fd);

        setPwBusy(false);

        if (res.ok) {
            setCurrentPassword("");
            setNewPassword("");
            setConfirmPassword("");
            setShowCurrentPassword(false);
            setShowNewPassword(false);
            setShowConfirmPassword(false);
            setPwStatus(
                Some(Ok(res.value.message || "Password changed successfully.")),
            );
        } else {
            setPwStatus(
                Some(Err(res.error.message || "Failed to change password.")),
            );
        }
    };

    const pwDisabled =
        pwBusy || !currentPassword || !newPassword || !confirmPassword;

    const status = pwStatus.some ? pwStatus.value : null;

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
                    <div className={`${styles.fieldRow} ${styles.fieldRowTop}`}>
                        <div className={styles.fieldText}>
                            <div className={styles.label}>Password</div>
                            <p className={styles.hint}>
                                Change your account password.
                            </p>
                        </div>

                        <div className={styles.control}>
                            <form
                                className={styles.passwordForm}
                                onSubmit={onChangePassword}
                            >
                                <div className={styles.passwordGrid}>
                                    <div className={styles.passwordField}>
                                        <label
                                            className={styles.inputLabel}
                                            htmlFor="current_password"
                                        >
                                            Current password
                                        </label>
                                        <div className={styles.inlineRow}>
                                            <input
                                                id="current_password"
                                                name="current_password"
                                                type={
                                                    showCurrentPassword
                                                        ? "text"
                                                        : "password"
                                                }
                                                className={styles.input}
                                                autoComplete="current-password"
                                                value={currentPassword}
                                                onChange={(e) => {
                                                    setPwStatus(None());
                                                    setCurrentPassword(
                                                        e.target.value,
                                                    );
                                                }}
                                                disabled={pwBusy}
                                            />
                                            <button
                                                type="button"
                                                className={`${styles.iconBtn} ${
                                                    showCurrentPassword
                                                        ? styles.iconBtnActive
                                                        : ""
                                                }`}
                                                onClick={() =>
                                                    setShowCurrentPassword(
                                                        (v) => !v,
                                                    )
                                                }
                                                aria-label={
                                                    showCurrentPassword
                                                        ? "Hide current password"
                                                        : "Show current password"
                                                }
                                                title={
                                                    showCurrentPassword
                                                        ? "Hide current password"
                                                        : "Show current password"
                                                }
                                                aria-pressed={
                                                    showCurrentPassword
                                                }
                                                disabled={pwBusy}
                                            >
                                                {showCurrentPassword ? (
                                                    <EyeOff
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                ) : (
                                                    <Eye
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    <div className={styles.passwordField}>
                                        <label
                                            className={styles.inputLabel}
                                            htmlFor="new_password"
                                        >
                                            New password
                                        </label>
                                        <div className={styles.inlineRow}>
                                            <input
                                                id="new_password"
                                                name="new_password"
                                                type={
                                                    showNewPassword
                                                        ? "text"
                                                        : "password"
                                                }
                                                className={styles.input}
                                                autoComplete="new-password"
                                                value={newPassword}
                                                onChange={(e) => {
                                                    setPwStatus(None());
                                                    setNewPassword(
                                                        e.target.value,
                                                    );
                                                }}
                                                disabled={pwBusy}
                                            />
                                            <button
                                                type="button"
                                                className={`${styles.iconBtn} ${
                                                    showNewPassword
                                                        ? styles.iconBtnActive
                                                        : ""
                                                }`}
                                                onClick={() =>
                                                    setShowNewPassword(
                                                        (v) => !v,
                                                    )
                                                }
                                                aria-label={
                                                    showNewPassword
                                                        ? "Hide new password"
                                                        : "Show new password"
                                                }
                                                title={
                                                    showNewPassword
                                                        ? "Hide new password"
                                                        : "Show new password"
                                                }
                                                aria-pressed={showNewPassword}
                                                disabled={pwBusy}
                                            >
                                                {showNewPassword ? (
                                                    <EyeOff
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                ) : (
                                                    <Eye
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    <div className={styles.passwordField}>
                                        <label
                                            className={styles.inputLabel}
                                            htmlFor="confirm_password"
                                        >
                                            Confirm new password
                                        </label>
                                        <div className={styles.inlineRow}>
                                            <input
                                                id="confirm_password"
                                                type={
                                                    showConfirmPassword
                                                        ? "text"
                                                        : "password"
                                                }
                                                className={styles.input}
                                                autoComplete="new-password"
                                                value={confirmPassword}
                                                onChange={(e) => {
                                                    setPwStatus(None());
                                                    setConfirmPassword(
                                                        e.target.value,
                                                    );
                                                }}
                                                disabled={pwBusy}
                                            />
                                            <button
                                                type="button"
                                                className={`${styles.iconBtn} ${
                                                    showConfirmPassword
                                                        ? styles.iconBtnActive
                                                        : ""
                                                }`}
                                                onClick={() =>
                                                    setShowConfirmPassword(
                                                        (v) => !v,
                                                    )
                                                }
                                                aria-label={
                                                    showConfirmPassword
                                                        ? "Hide confirm password"
                                                        : "Show confirm password"
                                                }
                                                title={
                                                    showConfirmPassword
                                                        ? "Hide confirm password"
                                                        : "Show confirm password"
                                                }
                                                aria-pressed={
                                                    showConfirmPassword
                                                }
                                                disabled={pwBusy}
                                            >
                                                {showConfirmPassword ? (
                                                    <EyeOff
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                ) : (
                                                    <Eye
                                                        className={
                                                            styles.iconBtnIcon
                                                        }
                                                        aria-hidden="true"
                                                    />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    <div className={styles.passwordActions}>
                                        <button
                                            type="submit"
                                            className={`${styles.button} ${styles.buttonInline} ${styles.primaryButton}`}
                                            disabled={pwDisabled}
                                        >
                                            {pwBusy
                                                ? "Changing..."
                                                : "Change password"}
                                        </button>
                                    </div>
                                </div>

                                {status && (
                                    <div
                                        className={`${styles.callout} ${
                                            status.ok
                                                ? styles.calloutOk
                                                : styles.calloutError
                                        }`}
                                        role={status.ok ? "status" : "alert"}
                                    >
                                        <div className={styles.calloutTitle}>
                                            {status.ok ? "Success" : "Error"}
                                        </div>
                                        <div className={styles.calloutBody}>
                                            {status.ok
                                                ? status.value
                                                : status.error}
                                        </div>
                                    </div>
                                )}
                            </form>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}
