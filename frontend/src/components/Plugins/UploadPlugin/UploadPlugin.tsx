import { useEffect, useRef, useState } from "react";
import styles from "./UploadPlugin.module.css";
import { Upload } from "lucide-react";
import { normalizeErrorMessage } from "@/components/Login/LoginView";
import PluginUploader from "@/api/plugins/pluginuploader";

export type UploadPluginProps = {
    onClose: () => void;
};

export default function UploadPlugin({ onClose }: UploadPluginProps) {
    const [link, setLink] = useState("");

    const [busy, setBusy] = useState(false);
    const [error, setError] = useState<{ title: string; body: string } | null>(
        null,
    );

    const linkRef = useRef<HTMLInputElement | null>(null);

    const title = "Upload Plugin";
    const subtitle =
        "Provide a link to a git repo to upload the plugin in that repo";
    const primaryLabel = "Upload Plugin";
    const PrimaryIcon = Upload;

    useEffect(() => {
        linkRef.current?.focus();

        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                e.preventDefault();
                e.stopPropagation();
                onClose();
            }
        };

        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, [onClose]);

    async function onSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);

        const u = link.trim();
        if (!u) {
            setError({
                title: "Missing Fields",
                body: "Git link is required",
            });
            return;
        }

        setBusy(true);
        const res = await PluginUploader(link);
        setBusy(false);
        if (res.some) {
            setError(
                normalizeErrorMessage(
                    res.value.message ?? "",
                    res.value.status,
                ),
            );
        } else {
            onClose();
        }
    }

    return (
        <div
            className={styles.viewerBackdrop}
            onClick={() => onClose()}
            role="dialog"
            aria-modal="true"
            aria-label={title}
        >
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>{title}</h1>
                    <p className={styles.subtle}>{subtitle}</p>
                </header>

                <form className={styles.form} onSubmit={onSubmit}>
                    <div className={styles.field}>
                        <label className={styles.label} htmlFor="link">
                            Git Link
                        </label>
                        <input
                            ref={linkRef}
                            id="link"
                            className={styles.input}
                            value={link}
                            onChange={(e) => setLink(e.target.value)}
                            autoComplete={"none"}
                            placeholder="Link"
                            disabled={busy}
                        />
                    </div>

                    {error && (
                        <div
                            className={`${styles.callout} ${
                                styles.calloutError
                            }`}
                            role={"alert"}
                        >
                            <div className={styles.calloutTitle}>
                                {error.title}
                            </div>
                            <div className={styles.calloutBody}>
                                {error.body}
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
