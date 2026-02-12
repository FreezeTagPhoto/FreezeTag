import styles from "./Pill.module.css";
import type React from "react";

type PillVariant = "menu" | "token" | "error";

export default function Pill({
    label,
    caret,
    variant = "menu",
    onClick,
    className = "",
    type = "button",
    invertCaret,
}: {
    label: React.ReactNode;
    caret?: boolean;
    variant?: PillVariant;
    onClick?: React.MouseEventHandler<HTMLButtonElement>;
    className?: string;
    type?: "button" | "submit" | "reset";
    invertCaret?: boolean;
}) {
    const variantClass =
        variant === "menu"
            ? styles.menu
            : variant === "token"
              ? styles.token
              : styles.error;

    return (
        <button
            type={type}
            className={`${styles.pill} ${variantClass} ${className} ${invertCaret ? styles.open : ""}`}
            onClick={onClick}
        >
            <span className={styles.label}>{label}</span>
            {caret && !invertCaret && <span className={styles.caret}>▾</span>}
            {caret && invertCaret && <span className={styles.caret}>▴</span>}
        </button>
    );
}