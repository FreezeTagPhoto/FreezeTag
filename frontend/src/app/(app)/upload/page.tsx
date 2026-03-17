"use client";
import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState, useEffect } from "react";
import styles from "./page.module.css";
import TagChangeButton from "@/components/UI/TagChangeButton/TagChangeButton";
import JobsHandler from "@/api/jobs/jobshandler";
import ProgressBar from "@/components/UI/ProgressBar/ProgressBar";
import FreezeTagSet from "@/common/freezetag/freezetagset";

const POLLING_DELAY = 500; // 0.5 seconds, in milliseconds

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

    const [selectedIds, setSelectedIds] = useState<FreezeTagSet<number>>(
        new FreezeTagSet(),
    );

    return (
        <div className={ids.length > 0 ? styles.page : styles.pageEmpty}>
            <div className={ids.length > 0 ? styles.toolbar : styles.none}>
                <FileUploadButton
                    job_id_callback={job_id_callback}
                    disabled={!!jobId}
                />
            </div>

            {progress > 0 && (
                <div className={styles.viewerBackdrop}>
                    <ProgressBar
                        progress={progress}
                        className={styles.progress_bar}
                    />
                </div>
            )}

            {ids.length > 0 && (
                <div className={styles.gallery_tags_container}>
                    <div className={styles.gallery_select_container}>
                        <div className={styles.select_container}>
                            <button
                                type="button"
                                className={styles.select_button}
                                onClick={() =>
                                    setSelectedIds(new FreezeTagSet(ids))
                                }
                            >
                                Select All
                            </button>
                            <button
                                type="button"
                                className={styles.select_button}
                                onClick={() =>
                                    setSelectedIds(new FreezeTagSet())
                                }
                            >
                                Deselect All
                            </button>
                        </div>
                        <div className={styles.gallery}>
                            <MassTaggingGallery
                                image_ids={ids}
                                selectedIds={selectedIds}
                                onChange={(id) => {
                                    setSelectedIds(selectedIds.toggle(id));
                                }}
                            />
                        </div>
                    </div>
                    <div className={styles.tag_change_container}>
                        <TagChangeButton image_ids={selectedIds} />
                    </div>
                </div>
            )}
        </div>
    );
}
