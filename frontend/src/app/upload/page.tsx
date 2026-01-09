"use client";
import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState } from "react";
import styles from "./page.module.css";
import TagChangeButton from "@/components/UI/TagChangeButton/TagChangeButton";

export default function Home() {
    const [ids, setIds] = useState<number[]>([]);
    const ids_retrieved_callback = (newIds: number[]) => {
        setIds(newIds);
    };
    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

    if (ids.length === 0) {
        return (
            <div className={styles.pageEmpty}>
                <FileUploadButton
                    ids_retrieved_callback={ids_retrieved_callback}
                />
            </div>
        );
    }

    return (
        <div className={styles.page}>
            <div className={styles.toolbar}>
                <FileUploadButton
                    ids_retrieved_callback={ids_retrieved_callback}
                />
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
