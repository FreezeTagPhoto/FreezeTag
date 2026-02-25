"use client";
import { useContext, useEffect, useState } from "react";
import styles from "./page.module.css";
import { UserContext } from "@/components/Auth/AuthGate";
import { UserHasPerm } from "@/api/permissions/permshelpers";
import { Ban, Clipboard, Info } from "lucide-react";
import { JobDetails, JobSummary } from "@/api/jobs/jobshelpers";
import JobsLister from "@/api/jobs/jobslister";
import ProgressBar from "@/components/UI/ProgressBar/ProgressBar";
import JobsCanceller from "@/api/jobs/jobscanceller";

const POLLING_FREQUENCY = 500; // milliseconds

export default function Home() {
    const [jobs, setJobs] = useState<JobSummary[]>([]);

    const [progress, setProgress] = useState<number[]>([]);

    const currentUser = useContext(UserContext);
    const userCanCancelJobs = UserHasPerm(currentUser, "write:jobs");

    const [jobDetails, setJobDetails] = useState<JobDetails | undefined>(
        undefined,
    );

    const JobsComparator = (a: JobSummary, b: JobSummary) => {
        const a_finished = a.status === "Finished";
        const b_finished = b.status === "Finished";

        const a_cancelled = a.status === "Cancelled";
        const b_cancelled = b.status === "Cancelled";

        const a_done = a_finished || a_cancelled;
        const b_done = b_finished || b_cancelled;

        const xnor = (a: boolean, b: boolean) => !(a !== b);

        if (xnor(a_done, b_done)) {
            if (xnor(a_finished, b_cancelled)) {
                if (a_finished) {
                    return 1;
                } else {
                    return -1;
                }
            }
            return a.uuid.localeCompare(b.uuid);
        } else if (a_done) {
            return 1;
        } else {
            return -1;
        }
    };

    useEffect(() => {
        const fetchJobs = async () => {
            const result = await JobsLister();
            if (result.ok) {
                const arr = result.value.sort(JobsComparator);
                const progressAmounts = arr.map(
                    (job) =>
                        (job.complete + job.errors) /
                        (job.complete + job.errors + job.in_progress),
                );
                setJobs(arr);
                setProgress(progressAmounts);
            } else {
                console.error(`Jobs Lister Error! ${result.error.message}`);
            }
        };
        setInterval(fetchJobs, POLLING_FREQUENCY);
    }, []);

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Job Management</h1>
            {jobs.length === 0 ? (
                <p>No jobs running!</p>
            ) : (
                <div className={styles.job_container}>
                    {jobs.map((job, idx) => (
                        <div key={job.uuid} className={styles.job_row}>
                            <div
                                className={`${styles.job_item} ${styles.job_name}`}
                                title={job.title}
                            >
                                <p className={styles.job_preview}>
                                    {job.title}
                                </p>
                            </div>
                            <button
                                type="button"
                                className={`${styles.job_item} ${styles.job_item_button}`}
                                onClick={() =>
                                    navigator.clipboard.writeText(job.uuid)
                                }
                                title="Copy UUID"
                            >
                                <Clipboard className={styles.icon} />
                                <p className={styles.job_item_label}>UUID</p>
                            </button>
                            <div
                                className={`${styles.job_item} ${styles.job_status}`}
                                title={job.status}
                            >
                                <Info
                                    className={`${styles.icon} ${job.status === "Cancelled" ? styles.job_status_cancelled : ""}`}
                                />
                                <p
                                    className={`${styles.job_preview} ${job.status === "Cancelled" ? styles.job_status_cancelled : ""}`}
                                >
                                    {job.status}
                                </p>
                            </div>
                            <div className={styles.job_item}>
                                <ProgressBar
                                    progress={progress[idx]}
                                    className={styles.job_progress}
                                    color={
                                        job.status === "Cancelled"
                                            ? "red"
                                            : job.status === "Finished"
                                              ? "green"
                                              : undefined
                                    }
                                />
                            </div>
                            <button
                                type="button"
                                className={`${styles.job_item} ${styles.job_item_button} ${styles.job_item_delete}`}
                                onClick={() => JobsCanceller(job.uuid)}
                                disabled={
                                    !userCanCancelJobs ||
                                    job.status === "Cancelled" ||
                                    job.status === "Finished"
                                }
                                title="Cancel Job"
                            >
                                <Ban className={styles.icon} />
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </main>
    );
}
