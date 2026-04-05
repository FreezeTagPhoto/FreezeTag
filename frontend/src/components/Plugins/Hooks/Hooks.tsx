import { useEffect } from "react";
import { AlarmClock, FishingHook, FileInput } from "lucide-react";
import styles from "./Hooks.module.css";
import { Plugin } from "@/api/plugins/pluginshelpers";

export type HooksProps = {
    onClose: () => void;
    plugin: Plugin;
};

export default function Hooks({ onClose, plugin }: HooksProps) {
    useEffect(() => {
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

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Hooks for {plugin.name}</h1>
                </header>
                <div className={styles.hooks_container}>
                    {Object.entries(plugin.hooks).map(([name, hook]) => (
                        <div key={name} className={`${styles.hooks_row}`}>
                            <div
                                className={`${styles.hooks_item} ${styles.text}`}
                                title={hook.friendly_name}
                            >
                                <FishingHook className={styles.icon} />
                                <p className={styles.text_preview}>
                                    {hook.friendly_name}
                                </p>
                            </div>
                            <div
                                className={`${styles.hooks_item} ${styles.text}`}
                                title={hook.signature}
                            >
                                <FileInput className={styles.icon} />
                                <p className={styles.text_preview}>
                                    {hook.signature}
                                </p>
                            </div>
                            <div
                                className={`${styles.hooks_item} ${styles.text}`}
                                title={hook.type}
                            >
                                <AlarmClock className={styles.icon} />
                                <p className={styles.text_preview}>
                                    {hook.type}
                                </p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
