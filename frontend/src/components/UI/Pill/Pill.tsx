import styles from "./Pill.module.css";
import type React from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

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

    const CaretIcon = invertCaret ? ChevronUp : ChevronDown;

    return (
        <button
            type={type}
            className={`${styles.pill} ${variantClass} ${className} ${
                invertCaret ? styles.open : ""
            }`}
            onClick={onClick}
        >
            <span className={styles.label}>{label}</span>

            {caret && (
                <span className={styles.caret} aria-hidden="true">
                    <CaretIcon className={styles.caretIcon} />
                </span>
            )}
        </button>
    );
}
