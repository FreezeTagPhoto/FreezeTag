"use client";

import { use } from "react";
import styles from "../page.module.css";
import AlbumDetailPage from "@/components/Albums/AlbumDetailPage/AlbumDetailPage";

type AlbumDetailRouteProps = {
    params: Promise<{
        albumName: string;
    }>;
};

export default function AlbumDetailRoute({ params }: AlbumDetailRouteProps) {
    const { albumName } = use(params);

    return (
        <main className={styles.main}>
            <AlbumDetailPage albumName={decodeURIComponent(albumName)} />
        </main>
    );
}
