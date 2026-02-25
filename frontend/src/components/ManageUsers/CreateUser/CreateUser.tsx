import { useState } from "react";
import styles from "./CreateUser.module.css";
import { Eye, EyeOff, UserPlus } from "lucide-react";
import UserCreator from "@/api/users/usercreator";
import { normalizeErrorMessage } from "@/components/Login/LoginView";
import PermsAdder from "@/api/permissions/permsadder";

export type CreateUserProps = {
    onClose: () => void;
};

export default function CreateUser({ onClose }: CreateUserProps) {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");

    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    const [busy, setBusy] = useState(false);
    const [error, setError] = useState<{ title: string; body: string } | null>(
        null,
    );
    const [success, setSuccess] = useState<string | null>(null);

    const title = "Create account";
    const subtitle = "Create a new FreezeTag user.";
    const primaryLabel = "Create User";
    const PrimaryIcon = UserPlus;

    async function onSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        setSuccess(null);

        const u = username.trim();
        if (!u || !password) {
            setError({
                title: "Missing Fields",
                body: "Username and password are required.",
            });
            return;
        }

        if (!confirmPassword) {
            setError({
                title: "Missing Fields",
                body: "Please confirm your password.",
            });
            return;
        }
        if (password !== confirmPassword) {
            setError({
                title: "Passwords Do Not Match",
                body: "Make sure both password fields are identical.",
            });
            return;
        }

        setBusy(true);
        try {
            const fd = new FormData();
            fd.set("username", u);
            fd.set("password", password);

            const res = await UserCreator(fd);
            if (!res.ok) {
                setError(
                    normalizeErrorMessage(
                        res.error.message ?? "",
                        res.error.status,
                    ),
                );
                return;
            } else {
                PermsAdder(res.value.id, [
                    "read:files",
                    "read:tags",
                    "read:plugins",
                ]);
                onClose();
            }
        } finally {
            setBusy(false);
        }
    }

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>{title}</h1>
                    <p className={styles.subtle}>{subtitle}</p>
                </header>
                <form className={styles.form} onSubmit={onSubmit}>
                    <div className={styles.field}>
                        <label className={styles.label} htmlFor="username">
                            Username
                        </label>
                        <input
                            id="username"
                            className={styles.input}
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            autoComplete={"new-username"}
                            placeholder="Username"
                            disabled={busy}
                        />
                    </div>

                    <div className={styles.field}>
                        <label className={styles.label} htmlFor="password">
                            Password
                        </label>

                        <div className={styles.inlineRow}>
                            <input
                                id="password"
                                className={styles.input}
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                type={showPassword ? "text" : "password"}
                                autoComplete={"new-password"}
                                placeholder={"Choose a password"}
                                disabled={busy}
                            />

                            <button
                                type="button"
                                className={`${styles.iconBtn} ${
                                    showPassword ? styles.iconBtnActive : ""
                                }`}
                                onClick={() => setShowPassword((v) => !v)}
                                aria-label={
                                    showPassword
                                        ? "Hide password"
                                        : "Show password"
                                }
                                title={
                                    showPassword
                                        ? "Hide password"
                                        : "Show password"
                                }
                                aria-pressed={showPassword}
                                disabled={busy}
                            >
                                {showPassword ? (
                                    <EyeOff
                                        className={styles.iconBtnIcon}
                                        aria-hidden="true"
                                    />
                                ) : (
                                    <Eye
                                        className={styles.iconBtnIcon}
                                        aria-hidden="true"
                                    />
                                )}
                            </button>
                        </div>
                    </div>

                    <div className={styles.field}>
                        <label
                            className={styles.label}
                            htmlFor="confirmPassword"
                        >
                            Confirm password
                        </label>

                        <div className={styles.inlineRow}>
                            <input
                                id="confirmPassword"
                                className={styles.input}
                                value={confirmPassword}
                                onChange={(e) =>
                                    setConfirmPassword(e.target.value)
                                }
                                type={showConfirmPassword ? "text" : "password"}
                                autoComplete="new-password"
                                placeholder="Re-enter password"
                                disabled={busy}
                            />

                            <button
                                type="button"
                                className={`${styles.iconBtn} ${
                                    showConfirmPassword
                                        ? styles.iconBtnActive
                                        : ""
                                }`}
                                onClick={() =>
                                    setShowConfirmPassword((v) => !v)
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
                                aria-pressed={showConfirmPassword}
                                disabled={busy}
                            >
                                {showConfirmPassword ? (
                                    <EyeOff
                                        className={styles.iconBtnIcon}
                                        aria-hidden="true"
                                    />
                                ) : (
                                    <Eye
                                        className={styles.iconBtnIcon}
                                        aria-hidden="true"
                                    />
                                )}
                            </button>
                        </div>
                    </div>

                    {(error || success) && (
                        <div
                            className={`${styles.callout} ${
                                error ? styles.calloutError : styles.calloutOk
                            }`}
                            role={error ? "alert" : "status"}
                        >
                            <div className={styles.calloutTitle}>
                                {error ? error.title : "Success"}
                            </div>
                            <div className={styles.calloutBody}>
                                {error ? error.body : success}
                            </div>
                        </div>
                    )}

                    <button
                        className={styles.primary}
                        type="submit"
                        disabled={busy}
                    >
                        <span className={styles.primaryInner}>
                            <PrimaryIcon
                                className={styles.primaryIcon}
                                aria-hidden="true"
                            />
                            <span>{busy ? "Working…" : primaryLabel}</span>
                        </span>
                    </button>
                </form>
            </div>
        </div>
    );
}
