import styles from "./ProgressBar.module.css";
export type ProgressBarProps = {
    progress: number;
    className?: string;
    color?: "red" | "green";
};

export default function ProgressBar(props: ProgressBarProps) {
    const colorClassName = `${props.color === "green" ? styles.progressBarGreen : props.color === "red" ? styles.progressBarRed : ""}`;
    return (
        <div>
            <label htmlFor="job-progress" className={styles.hidden}>
                Job Progress:
            </label>
            <progress
                id="job-progress"
                value={props.progress}
                className={`${styles.progressBar} ${colorClassName} ${props.className ?? ""}`}
            >
                {props.progress * 100}%
            </progress>
        </div>
    );
}
