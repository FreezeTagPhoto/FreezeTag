"use client";
import SearchHandler from "@/api/query/searchhandler";
import styles from "./page.module.css";
import Gallery from "@/components/Gallery/Gallery";
import TopBar from "@/components/TopBar/TopBar";
import { useEffect, useState } from "react";

export default function Home() {
  const [images_ids, set_image_ids] = useState<number[]>([]);

  const onChangeHandler = async (query: string) => {
    const result = await SearchHandler(query);
    if (result.ok) {
      set_image_ids(result.value);
    } else {
      console.log(
        "Got " +
          result.error.status +
          " from backend with message " +
          result.error.message,
      );
    }
  };

  useEffect(() => {
    onChangeHandler("");
  }, []);

  return (
    <>
      <TopBar onChangeHandler={onChangeHandler} />
      <main className={styles.main}>
        <header className={styles.headerRow}>
          <div>
            <h1 className={styles.h1}>Gallery</h1>
            <p className={styles.subtle}>
              {images_ids.length} {images_ids.length !== 1 ? "images" : "image"}
            </p>
          </div>
          <div className={styles.pillsRow} />
        </header>

        <Gallery image_ids={images_ids} />
      </main>
    </>
  );
}
