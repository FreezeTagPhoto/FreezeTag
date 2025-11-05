import styles from "./Pill.module.css";

export default function Pill({
  label,
  caret,
}: {
  label: string;
  caret?: boolean;
}) {
  return (
    <button className={styles.pill}>
      <span>{label}</span>
      {caret && <span className={styles.caret}>▾</span>}
    </button>
  );
}
