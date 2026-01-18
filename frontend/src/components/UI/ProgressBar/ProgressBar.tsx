import styles from "./ProgressBar.module.css";
export type ProgressBarProps = {
    progress: number;
};

export default function ProgressBar(props: ProgressBarProps) {
    return (
        <div className={styles.viewerBackdrop}>
            <label htmlFor="upload-progress" className={styles.hidden}>
                Upload Progress:
            </label>
            <progress
                id="upload-progress"
                value={props.progress}
                className={styles.progressBar}
            >
                {props.progress * 100}%
            </progress>
        </div>
    );
}
