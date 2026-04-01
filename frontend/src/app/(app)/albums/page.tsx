"use client";

import styles from "./page.module.css";
import AlbumsPage from "@/components/Albums/AlbumPage.tsx/AlbumPage";

export default function Home() {
    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Albums</h1>
            <AlbumsPage />
        </main>
    );
}