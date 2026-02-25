import styles from "./ProgressBar.module.css";
export type ProgressBarProps = {
    progress: number;
    className?: string;
};

export default function ProgressBar(props: ProgressBarProps) {
    return (
        <div>
            <label htmlFor="job-progress" className={styles.hidden}>
                Job Progress:
            </label>
            <progress
                id="job-progress"
                value={props.progress}
                className={`${styles.progressBar} ${props.className ?? ""}`}
            >
                {props.progress * 100}%
            </progress>
        </div>
    );
}
