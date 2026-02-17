"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import logo from "@/icons/freezetag+text.svg";

import LoginHandler from "@/api/auth/loginhandler";
import UserCreator from "@/api/users/usercreator";

import styles from "./LoginView.module.css";
import { Eye, EyeOff, LogIn, UserPlus } from "lucide-react";

type OptionSome<T> = { kind: "some"; value: T };
type OptionNone = { kind: "none" };

function isOptionSome<T>(x: unknown): x is OptionSome<T> {
    if (!x || typeof x !== "object") return false;
    const o = x as Record<string, unknown>;
    return o.kind === "some" && "value" in o;
}

function isOptionNone(x: unknown): x is OptionNone {
    if (!x || typeof x !== "object") return false;
    const o = x as Record<string, unknown>;
    return o.kind === "none";
}

function optionToValue<T>(opt: unknown): T | null {
    if (isOptionSome<T>(opt)) return opt.value;
    if (isOptionNone(opt)) return null;

    if (opt && typeof opt === "object") {
        const o = opt as Record<string, unknown>;
        if ("value" in o) return o.value as T;
    }
    return null;
}

function toTitle(s: string) {
    const small = new Set([
        "a",
        "an",
        "and",
        "as",
        "at",
        "but",
        "by",
        "for",
        "from",
        "if",
        "in",
        "into",
        "nor",
        "of",
        "on",
        "or",
        "over",
        "per",
        "the",
        "to",
        "vs",
        "via",
        "with",
        "without",
    ]);

    const words = s.trim().replace(/\s+/g, " ").split(" ").filter(Boolean);

    return words
        .map((w, i) => {
            const lower = w.toLowerCase();
            if (i !== 0 && small.has(lower)) return lower;
            return lower.length
                ? lower[0].toUpperCase() + lower.slice(1)
                : lower;
        })
        .join(" ");
}

function splitAuthMessage(raw: string): { title: string; body: string } {
    const msg = raw.trim().replace(/^"+|"+$/g, "");
    const parts = msg.split(":");
    if (parts.length >= 2) {
        const title = toTitle(parts[0]);
        const body = parts.slice(1).join(":").trim();
        const bodyNorm = body.length
            ? body[0].toUpperCase() + body.slice(1)
            : body;
        return { title, body: bodyNorm };
    }
    return { title: "Error", body: msg };
}

function normalizeErrorMessage(raw: string, status?: number) {
    const fallback = { title: "Error", body: "Something went wrong." };

    const parsedJsonInner = (s: string) => {
        const t = s.trim();
        if (!t.startsWith("{")) return null;
        try {
            const obj = JSON.parse(t) as { error?: unknown; message?: unknown };
            const inner =
                typeof obj.error === "string"
                    ? obj.error.trim()
                    : typeof obj.message === "string"
                      ? obj.message.trim()
                      : "";
            return inner || null;
        } catch {
            return null;
        }
    };

    const trimmed = (raw ?? "").trim();
    const inner = parsedJsonInner(trimmed);

    let out: { title: string; body: string };

    if (status === 401 || status === 403) {
        out = splitAuthMessage(
            "authentication failed: invalid username or password",
        );
    } else if (inner) {
        out = splitAuthMessage(inner);
    } else if (trimmed) {
        out = splitAuthMessage(trimmed);
    } else {
        out = fallback;
    }

    return { title: toTitle(out.title), body: out.body };
}

type Mode = "login" | "create";

export function Logo() {
    return (
        <div className={styles.logoWrap} aria-label="FreezeTag">
            <Image
                src={logo}
                alt="FreezeTag"
                fill
                priority
                sizes="420px"
                className={styles.logoImg}
            />
        </div>
    );
}

export default function LoginView({ mode }: { mode: Mode }) {
    const router = useRouter();

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

    const title = mode === "login" ? "Sign in" : "Create account";
    const subtitle =
        mode === "login"
            ? "Use your FreezeTag credentials to continue."
            : "Create a new FreezeTag user.";
    const primaryLabel = mode === "login" ? "Sign in" : "Create user";
    const PrimaryIcon = mode === "login" ? LogIn : UserPlus;

    const footer = useMemo(() => {
        if (mode === "login") {
            return {
                text: "No account?",
                href: "/login/createuser",
                link: "Create one",
            };
        }
        return {
            text: "Already have an account?",
            href: "/login",
            link: "Sign in",
        };
    }, [mode]);

    async function doLogin(u: string, p: string) {
        const fd = new FormData();
        fd.set("username", u);
        fd.set("password", p);

        const optErr = await LoginHandler(fd);
        const errVal = optionToValue<{ status: number; message: string }>(
            optErr,
        );

        if (errVal) {
            setError(
                normalizeErrorMessage(errVal.message ?? "", errVal.status),
            );
            return false;
        }

        router.replace("/");
        router.refresh();
        return true;
    }

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

        if (mode === "create") {
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
        }

        setBusy(true);
        try {
            if (mode === "login") {
                await doLogin(u, password);
                return;
            }

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
            }

            const loggedIn = await doLogin(u, password);
            if (!loggedIn) {
                setSuccess(
                    `User "${res.value.username}" created. Please sign in.`,
                );
            }
        } finally {
            setBusy(false);
        }
    }

    return (
        <section className={styles.card} aria-label={title}>
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
                        autoComplete={
                            mode === "login" ? "username" : "new-username"
                        }
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
                            autoComplete={
                                mode === "login"
                                    ? "current-password"
                                    : "new-password"
                            }
                            placeholder={
                                mode === "login"
                                    ? "Your password"
                                    : "Choose a password"
                            }
                            disabled={busy}
                        />

                        <button
                            type="button"
                            className={`${styles.iconBtn} ${
                                showPassword ? styles.iconBtnActive : ""
                            }`}
                            onClick={() => setShowPassword((v) => !v)}
                            aria-label={
                                showPassword ? "Hide password" : "Show password"
                            }
                            title={
                                showPassword ? "Hide password" : "Show password"
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

                {mode === "create" && (
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
                )}

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

                <div className={styles.footerRow}>
                    <span className={styles.footerText}>{footer.text}</span>
                    <Link className={styles.link} href={footer.href}>
                        {footer.link}
                    </Link>
                </div>
            </form>
        </section>
    );
}
