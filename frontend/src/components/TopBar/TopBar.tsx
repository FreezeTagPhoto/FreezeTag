"use client";
import styles from "./TopBar.module.css";
import Pill from "@/components/UI/Pill";

export default function TopBar() {
  return (
    <div className={styles.bar}>
      <div className={styles.searchWrap}>
        <span className={styles.searchIcon} aria-hidden>
          🔍
        </span>
        <input
          className={styles.search}
          placeholder="Search…"
          aria-label="Search"
        />
        <button className={styles.clear} aria-label="Clear">
          ✕
        </button>
      </div>

      <div className={styles.pills}>
        <Pill label="Tags" caret />
        <Pill label="People" caret />
        <Pill label="Export" caret />
      </div>
    </div>
  );
}
