"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { FolderOpen } from "lucide-react";
import AlbumLister, { AlbumItem } from "@/api/albums/albumlister";
import styles from "./AlbumPage.module.css";

export default function AlbumsPage() {
    const [albums, setAlbums] = useState<AlbumItem[] | null>(null);
    const [error, setError] = useState<string>("");

    useEffect(() => {
        const loadAlbums = async () => {
            const result = await AlbumLister();
            if (!result.ok) {
                setError(result.error.message || "Failed to load albums.");
            } else {
                setError("");
                setAlbums(result.value);
            }
        };

        loadAlbums();
    }, []);

    return (
        <section aria-label="Albums Management" className={styles.wrap}>
            {error && <p className={styles.error}>{error}</p>}
            {!albums || albums.length === 0 ? (
                <div className={styles.empty}>
                    <p className={styles.subtle}>No albums found.</p>
                </div>
            ) : (
                <ul className={styles.grid}>
                    {albums.map((album) => (
                        <li key={album.id} className={styles.item}>
                            <Link
                                href={`/albums/${encodeURIComponent(album.id.toString())}`}
                                className={styles.card}
                            >
                                <FolderOpen className={styles.icon} />
                                <span className={styles.name}>
                                    {album.name}
                                </span>
                            </Link>
                        </li>
                    ))}
                </ul>
            )}
        </section>
    );
}
