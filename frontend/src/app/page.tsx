import styles from "./page.module.css";
import Gallery from "@/components/Gallery/Gallery";
import TopBar from "@/components/TopBar/TopBar";

export default function Home() {
  const images_ids: string | number[] = [];
  return (
    <>
      <TopBar />
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
