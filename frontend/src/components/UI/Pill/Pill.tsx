import styles from "./Pill.module.css";

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
    label: string;
    caret?: boolean;
    variant?: PillVariant;
    onClick?: () => void;
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
