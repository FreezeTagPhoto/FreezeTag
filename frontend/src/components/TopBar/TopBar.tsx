"use client";
import { useState } from "react";
import styles from "./TopBar.module.css";
import Pill from "@/components/UI/Pill/Pill";

type TopBarProps = {
  onChangeHandler: (value: string) => void;
};

export default function TopBar({ onChangeHandler }: TopBarProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    onChangeHandler(value);
  };

  const handleClear = () => {
    handleSearchChange("");
  };

  return (
    <div className={styles.bar}>
      <div className={styles.searchWrap}>
        <span className={styles.searchIcon} aria-hidden>
          🔍
        </span>
        <input
          className={styles.search}
          placeholder="Search..."
          aria-label="Search"
          value={searchTerm}
          onChange={(e) => handleSearchChange(e.target.value)}
        />
        <button
          className={styles.clear}
          aria-label="Clear"
          onClick={handleClear}
          type="button"
        >
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
