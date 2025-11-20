"use client";
import Gallery from "@/components/Gallery/Gallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState } from "react";
import styles from "./page.module.css";

export default function Home() {
  const [ids, setIds] = useState<number[]>([]);
  const ids_retrieved_callback = (ids: number[]) => {
    setIds(ids);
  };
  return (
    <div className={styles.page}>
      <div className={styles.center}>
        <FileUploadButton ids_retrieved_callback={ids_retrieved_callback} />
      </div>

      {ids.length > 0 && (
        <div className={styles.gallery}>
          <Gallery image_ids={ids} />
        </div>
      )}
    </div>
  );
}
