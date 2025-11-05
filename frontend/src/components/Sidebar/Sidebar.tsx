import Image from "next/image";
import logoUrl from "@/app/freezetag+text.svg";
import styles from "./Sidebar.module.css";

export default function Sidebar() {
  return (
    <div className={styles.wrap}>
      <div className={styles.brand}>
        <div className={styles.logoWrap}>
          <Image
            src={logoUrl}
            alt="FreezeTag"
            fill
            priority
            sizes="240px"
            className={styles.logoImg}
          />
        </div>
      </div>

      <nav className={styles.menu}>
        <button className={styles.item}>Manage</button>
        <button className={styles.item}>Settings</button>
        <button className={styles.item}>Plugins</button>
        <button className={styles.item}>Accounts</button>
      </nav>
    </div>
  );
}
