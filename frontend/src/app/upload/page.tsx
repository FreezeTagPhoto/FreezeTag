"use client";
import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState, useEffect } from "react";
import styles from "./page.module.css";
import TagChangeButton from "@/components/UI/TagChangeButton/TagChangeButton";
import JobsHandler from "@/api/jobs/jobshandler";

const POLLING_DELAY = 200; // 0.2 seconds, in milliseconds

export default function Home() {
    const [ids, setIds] = useState<number[]>([]);
    const [jobId, setJobId] = useState<string>("");
    const [progress, setProgress] = useState<number>(-1);

    const job_id_callback = (newJobId: string) => {
        setJobId(newJobId);
    };

    useEffect(() => {
        if (jobId == "") {
            return;
        }
        const interval_id = setInterval(async () => {
            const result = await JobsHandler(jobId);
            if (!result.ok) {
                console.error(
                    `Error from jobs api! ${result.error.status} ${result.error.message}`,
                );
                return;
            }
            if (!result.value.ok) {
                console.log(`Progress: ${result.value.error}`);
                setProgress(result.value.error);
                return;
            }
            const image_map = result.value.value;
            const image_ids = [];
            for (const [key, image] of image_map) {
                if (image.ok) {
                    image_ids.push(image.value);
                } else {
                    console.error(
                        `Error for image ${key} with message ${image.error}`,
                    );
                }
            }
            clearInterval(interval_id);
            setJobId("");
            setProgress(-1);
            setIds(image_ids);
        }, POLLING_DELAY);
    }, [jobId]);

    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

    if (ids.length === 0) {
        return (
            <div className={styles.pageEmpty}>
                <FileUploadButton job_id_callback={job_id_callback} />
            </div>
        );
    }

    return (
        <div className={styles.page}>
            <div className={styles.toolbar}>
                <FileUploadButton job_id_callback={job_id_callback} />
            </div>

            <div className={styles.gallery_tags_container}>
                <div className={styles.gallery}>
                    <MassTaggingGallery
                        image_ids={ids}
                        onChange={(ids) => setSelectedIds(ids)}
                    />
                </div>
                <TagChangeButton image_ids={selectedIds} />
            </div>
        </div>
    );
}
