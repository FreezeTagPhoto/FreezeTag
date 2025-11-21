"use client";
import Gallery from "@/components/Gallery/Gallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState } from "react";
import styles from "./page.module.css";

export default function Home() {
  const [ids, setIds] = useState<number[]>([]);
  const ids_retrieved_callback = (newIds: number[]) => {
    setIds(newIds);
  };

  if (ids.length === 0) {
    return (
      <div className={styles.pageEmpty}>
        <FileUploadButton ids_retrieved_callback={ids_retrieved_callback} />
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <FileUploadButton ids_retrieved_callback={ids_retrieved_callback} />
      </div>

      <div className={styles.gallery}>
        <Gallery image_ids={ids} />
      </div>
    </div>
  );
}
