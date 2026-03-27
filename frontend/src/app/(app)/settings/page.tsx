"use client";

import { useContext, useEffect, useRef, useState } from "react";
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
import ProfilePictureGetter from "@/api/users/profilepicturegetter";
import ProfilePictureSetter from "@/api/users/profilepicturesetter";
import ProfilePictureReset from "@/api/users/profilepicturereset";
import { UserContext } from "@/components/Auth/AuthGate";
import { ProfilePictureContext } from "@/components/Auth/ProfilePictureContext";
import { Option, Some, None } from "@/common/option";
import { Result, Ok, Err } from "@/common/result";
import { Camera, Eye, EyeOff, User, RotateCcw } from "lucide-react";

type ThemeName =
    | (typeof LightThemeRegistry)[number]
    | (typeof DarkThemeRegistry)[number];

type PwKey = "current" | "new" | "confirm";

const PasswordField = ({
    id,
    label,
    value,
    onChange,
    show,
    onToggleShow,
    autoComplete,
    disabled,
}: {
    id: string;
    label: string;
    value: string;
    onChange: (v: string) => void;
    show: boolean;
    onToggleShow: () => void;
    autoComplete: string;
    disabled: boolean;
}) => (
    <div className={styles.passwordField}>
        <label className={styles.inputLabel} htmlFor={id}>
            {label}
        </label>
        <div className={styles.inlineRow}>
            <input
                id={id}
                name={id}
                type={show ? "text" : "password"}
                className={styles.input}
                autoComplete={autoComplete}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                disabled={disabled}
            />
            <button
                type="button"
                className={`${styles.iconBtn} ${
                    show ? styles.iconBtnActive : ""
                }`}
                onClick={onToggleShow}
                aria-label={show ? `Hide ${label}` : `Show ${label}`}
                title={show ? `Hide ${label}` : `Show ${label}`}
                aria-pressed={show}
                disabled={disabled}
            >
                {show ? (
                    <EyeOff className={styles.iconBtnIcon} aria-hidden />
                ) : (
                    <Eye className={styles.iconBtnIcon} aria-hidden />
                )}
            </button>
        </div>
    </div>
);

export default function SettingsPage() {
    const user = useContext(UserContext);
    const { refreshProfilePicture } = useContext(ProfilePictureContext);

    const fileInputRef = useRef<HTMLInputElement>(null);
    const [avatarUrl, setAvatarUrl] = useState<string | null>(null);
    const [avatarBusy, setAvatarBusy] = useState(false);
    const [avatarStatus, setAvatarStatus] =
        useState<Option<Result<string, string>>>(None());
    useEffect(() => {
        if (!user) return;
        ProfilePictureGetter(user.user_id).then((result) => {
            if (result.ok) {
                const url = URL.createObjectURL(result.value);
                setAvatarUrl((prev) => {
                    if (prev) URL.revokeObjectURL(prev);
                    return url;
                });
            }
        });
    }, [user]);

    useEffect(() => {
        return () => {
            if (avatarUrl) URL.revokeObjectURL(avatarUrl);
        };
    }, []);

    const handleAvatarFileChange = async (
        e: React.ChangeEvent<HTMLInputElement>,
    ) => {
        const file = e.target.files?.[0];

        // reset the input so the same file can be re-selected after an error
        if (fileInputRef.current) fileInputRef.current.value = "";
        if (!file || !user) return;

        setAvatarBusy(true);
        setAvatarStatus(None());

        const result = await ProfilePictureSetter(user.user_id, file);

        if (result.ok) {
            const refreshed = await ProfilePictureGetter(user.user_id);
            if (refreshed.ok) {
                const url = URL.createObjectURL(refreshed.value);
                setAvatarUrl((prev) => {
                    if (prev) URL.revokeObjectURL(prev);
                    return url;
                });
            }
            refreshProfilePicture();
            setAvatarStatus(Some(Ok("Profile picture updated.")));
        } else {
            setAvatarStatus(
                Some(
                    Err(
                        result.error.message ||
                            "Failed to update profile picture.",
                    ),
                ),
            );
        }

        setAvatarBusy(false);
    };

    const handleResetProfilePicture = async () => {
        if (!user) return;

        setAvatarBusy(true);
        setAvatarStatus(None());

        const result = await ProfilePictureReset(user.user_id);

        if (result.ok) {
            const refreshed = await ProfilePictureGetter(user.user_id);
            if (refreshed.ok) {
                const url = URL.createObjectURL(refreshed.value);
                setAvatarUrl((prev) => {
                    if (prev) URL.revokeObjectURL(prev);
                    return url;
                });
            }
            refreshProfilePicture();
            setAvatarStatus(Some(Ok("Profile picture reset to default.")));
        } else {
            setAvatarStatus(
                Some(
                    Err(
                        result.error.message ||
                            "Failed to reset profile picture.",
                    ),
                ),
            );
        }

        setAvatarBusy(false);
    };

    const [theme, setTheme] = useState<ThemeName>("Catppuccin Mocha");
    const [units, setUnits] = useState<UnitSystem>("metric");

    const [isEditingPassword, setIsEditingPassword] = useState(false);

    const [pw, setPw] = useState({ current: "", next: "", confirm: "" });
    const [pwBusy, setPwBusy] = useState(false);
    const [pwStatus, setPwStatus] =
        useState<Option<Result<string, string>>>(None());
    const [pwShow, setPwShow] = useState({
        current: false,
        new: false,
        confirm: false,
    });

    useEffect(() => {
        const initialTheme = ThemeGetter() as ThemeName;
        setTheme(initialTheme);
        ApplyTheme(initialTheme);

        const initialUnits = UnitsGetter();
        setUnits(initialUnits);
        ApplyUnits(initialUnits);
    }, []);

    const resetPw = () => {
        setPw({ current: "", next: "", confirm: "" });
        setPwShow({ current: false, new: false, confirm: false });
    };

    const handleToggleEdit = () => {
        resetPw();
        setPwStatus(None());
        setIsEditingPassword((prev) => !prev);
    };

    const setPwField = (key: keyof typeof pw, value: string) => {
        setPwStatus(None());
        setPw((prev) => ({ ...prev, [key]: value }));
    };

    const toggleShow = (key: PwKey) =>
        setPwShow((prev) => ({ ...prev, [key]: !prev[key] }));

    const onChangePassword: React.FormEventHandler<HTMLFormElement> = async (
        e,
    ) => {
        e.preventDefault();
        setPwStatus(None());

        if (!pw.current || !pw.next || !pw.confirm) {
            setPwStatus(Some(Err("Please fill out all fields.")));
            return;
        }

        if (pw.next !== pw.confirm) {
            setPwStatus(Some(Err("New passwords do not match.")));
            return;
        }

        if (pw.next === pw.current) {
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
        fd.set("current_password", pw.current);
        fd.set("new_password", pw.next);

        const res = await PasswordChanger(fd);

        setPwBusy(false);

        if (res.ok) {
            resetPw();
            setIsEditingPassword(false);
            // setPwStatus(
            //     Some(Ok(res.value.message || "Password changed successfully.")),
            // );
        } else {
            setPwStatus(
                Some(Err(res.error.message || "Failed to change password.")),
            );
        }
    };

    const pwDisabled = pwBusy || !pw.current || !pw.next || !pw.confirm;
    const status = pwStatus.some ? pwStatus.value : null;

    return (
        <main className={styles.main}>
            <header className={styles.header}>
                <h1 className={styles.h1}>Settings</h1>
                <p className={styles.subtitle}>
                    Customize preferences and manage account security.
                </p>
            </header>

            <section
                id="profile"
                className={styles.panel}
                aria-labelledby="profile-heading"
            >
                <div className={styles.sectionHeader}>
                    <h2 id="profile-heading" className={styles.sectionTitle}>
                        Profile
                    </h2>
                    <p className={styles.sectionDescription}>
                        Profile detail management.
                    </p>
                </div>

                <div className={styles.fields}>
                    <div className={`${styles.fieldRow} ${styles.fieldRowTop}`}>
                        <div className={styles.fieldText}>
                            <div className={styles.label}>Profile picture</div>
                            <p className={styles.hint}>
                                Accepted formats: JPEG, PNG, WebP, HEIC, AVIF, TIFF, SVG.
                            </p>
                        </div>

                        <div className={styles.control}>
                            <div className={styles.avatarInnerGrid}>
                                <div
                                    className={styles.avatarRing}
                                    aria-hidden="true"
                                >
                                    {avatarUrl ? (
                                        <img
                                            src={avatarUrl}
                                            alt="Your profile picture"
                                            className={styles.avatarImg}
                                        />
                                    ) : (
                                        <div className={styles.avatarFallback}>
                                            <User
                                                className={
                                                    styles.avatarFallbackIcon
                                                }
                                                aria-hidden
                                            />
                                        </div>
                                    )}
                                </div>

                                <div className={styles.avatarActions}>
                                    <input
                                        ref={fileInputRef}
                                        type="file"
                                        accept="image/jpeg,image/png,image/webp,image/heic,image/heif,image/avif,image/tiff,image/svg+xml"
                                        className={styles.hiddenFileInput}
                                        aria-label="Upload profile picture"
                                        disabled={avatarBusy}
                                        onChange={handleAvatarFileChange}
                                    />
                                    <div className={styles.avatarBtnsGroup}>
                                        <div className={styles.avatarBtns}>
                                            <button
                                                type="button"
                                                className={`${styles.button} ${styles.buttonInline} ${styles.avatarUploadBtn}`}
                                                disabled={avatarBusy}
                                                onClick={() =>
                                                    fileInputRef.current?.click()
                                                }
                                            >
                                                <Camera
                                                    className={
                                                        styles.btnLeadIcon
                                                    }
                                                    aria-hidden
                                                />
                                                {avatarBusy
                                                    ? "Uploading…"
                                                    : "Change profile picture"}
                                            </button>
                                            <button
                                                type="button"
                                                className={`${styles.iconBtn}`}
                                                disabled={avatarBusy}
                                                onClick={
                                                    handleResetProfilePicture
                                                }
                                                title="Reset to default profile picture"
                                                aria-label="Reset to default profile picture"
                                            >
                                                <RotateCcw
                                                    className={
                                                        styles.iconBtnIcon
                                                    }
                                                    aria-hidden
                                                />
                                            </button>
                                        </div>

                                        {avatarStatus.some && (
                                            <div
                                                className={`${styles.callout} ${
                                                    avatarStatus.value.ok
                                                        ? styles.calloutOk
                                                        : styles.calloutError
                                                }`}
                                                role={
                                                    avatarStatus.value.ok
                                                        ? "status"
                                                        : "alert"
                                                }
                                            >
                                                <div
                                                    className={
                                                        styles.calloutTitle
                                                    }
                                                >
                                                    {avatarStatus.value.ok
                                                        ? "Updated"
                                                        : "Error"}
                                                </div>
                                                <div
                                                    className={
                                                        styles.calloutBody
                                                    }
                                                >
                                                    {avatarStatus.value.ok
                                                        ? "Profile picture updated."
                                                        : avatarStatus.value
                                                              .error}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

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
                                className={`${styles.button} ${styles.buttonInline} ${
                                    isEditingPassword ? styles.cancelButton : ""
                                }`}
                                onClick={handleToggleEdit}
                            >
                                {isEditingPassword
                                    ? "Cancel"
                                    : "Change password"}
                            </button>
                        </div>
                    </div>

                    {isEditingPassword && (
                        <div className={styles.expandedForm}>
                            <form
                                className={styles.passwordForm}
                                onSubmit={onChangePassword}
                            >
                                <div className={styles.passwordGrid}>
                                    <div className={styles.fullWidthField}>
                                        <PasswordField
                                            id="current_password"
                                            label="Current password"
                                            value={pw.current}
                                            onChange={(v) =>
                                                setPwField("current", v)
                                            }
                                            show={pwShow.current}
                                            onToggleShow={() =>
                                                toggleShow("current")
                                            }
                                            autoComplete="current-password"
                                            disabled={pwBusy}
                                        />
                                    </div>

                                    <PasswordField
                                        id="new_password"
                                        label="New password"
                                        value={pw.next}
                                        onChange={(v) => setPwField("next", v)}
                                        show={pwShow.new}
                                        onToggleShow={() => toggleShow("new")}
                                        autoComplete="new-password"
                                        disabled={pwBusy}
                                    />

                                    <PasswordField
                                        id="confirm_password"
                                        label="Confirm new password"
                                        value={pw.confirm}
                                        onChange={(v) =>
                                            setPwField("confirm", v)
                                        }
                                        show={pwShow.confirm}
                                        onToggleShow={() =>
                                            toggleShow("confirm")
                                        }
                                        autoComplete="new-password"
                                        disabled={pwBusy}
                                    />

                                    <div
                                        className={`${styles.passwordActions} ${styles.fullWidthField}`}
                                    >
                                        <button
                                            type="submit"
                                            className={`${styles.button} ${styles.buttonInline} ${styles.primaryButton}`}
                                            disabled={pwDisabled}
                                        >
                                            {pwBusy
                                                ? "Saving..."
                                                : "Save new password"}
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
                    )}
                </div>
            </section>
        </main>
    );
}
