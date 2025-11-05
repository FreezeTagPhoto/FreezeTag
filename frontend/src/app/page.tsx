import styles from "./page.module.css";
import Sidebar from "@/components/Sidebar/Sidebar";
import TopBar from "@/components/TopBar/TopBar";
import Gallery from "@/components/Gallery/Gallery";

export default function Home() {
  return (
    <div className={styles.shell}>
      <aside className={styles.nav}>
        <Sidebar />
      </aside>

      <div className={styles.content}>
        <TopBar />
        <main className={styles.main}>
          <header className={styles.headerRow}>
            <div>
              <h1 className={styles.h1}>Gallery</h1>
              <p className={styles.subtle}>100 images found</p>
            </div>
            <div className={styles.pillsRow}>
            </div>
          </header>
          <Gallery />
        </main>
      </div>
    </div>
  );
}
