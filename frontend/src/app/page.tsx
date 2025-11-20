import styles from "./page.module.css";
import Gallery from "@/components/Gallery/Gallery";
import TopBar from "@/components/TopBar/TopBar";

export default function Home() {
  return (
    <>
      <TopBar />
      <main className={styles.main}>
        <header className={styles.headerRow}>
          <div>
            <h1 className={styles.h1}>Gallery</h1>
            <p className={styles.subtle}>100 images</p>
          </div>
          <div className={styles.pillsRow} />
        </header>

        <Gallery image_ids={[]} />
      </main>
    </>
  );
}
