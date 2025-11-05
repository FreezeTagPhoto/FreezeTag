import styles from "./page.module.css";
import Gallery from "@/components/Gallery/Gallery";

export default function Home() {
  return (
    <main className={styles.main}>
      <header className={styles.headerRow}>
        <div>
          <h1 className={styles.h1}>Gallery</h1>
          <p className={styles.subtle}>100 images found</p>
        </div>
        <div className={styles.pillsRow} />
      </header>

      <Gallery />
    </main>
  );
}
