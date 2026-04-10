"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { FolderOpen, Globe, Lock } from "lucide-react";
import AlbumLister, { AlbumItem } from "@/api/albums/albumlister";
import styles from "./AlbumPage.module.css";
import AlbumCreator from "@/api/albums/albumcreator";

export default function AlbumsPage() {
    const [albums, setAlbums] = useState<AlbumItem[] | null>(null);
    const [error, setError] = useState<string>("");
    const [albumPopup, setAlbumPopup] = useState<boolean>(false);

    const loadAlbums = async () => {
        const result = await AlbumLister();
        if (!result.ok) {
            setError(result.error.message || "Failed to load albums.");
        } else {
            setError("");
            setAlbums(result.value);
        }
    };

    useEffect(() => {
        loadAlbums();
    }, []);

    const handleCreateAlbum = async (name: string, publicState: number) => {
        const result = await AlbumCreator(name, publicState);
        if (result.ok) {
            loadAlbums();
            setAlbumPopup(false);
        } else {
            setError(result.error.message || "Failed to create album.");
        }

    };

    return (
        <section aria-label="Albums Management" className={styles.wrap}>
            <div className={styles.add}>
                <button className={styles.button} onClick={() => setAlbumPopup(true)}>
                    New Album
                </button>
            </div>
            {albumPopup && (
                <AlbumNewPopup
                    onClose={() => {setAlbumPopup(false); setError("");}}
                    onSubmit={handleCreateAlbum}
                    error={error}
                />
            )}
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
                                <span className={styles.name}>{album.name}</span>
                            </Link>
                        </li>
                    ))}
                </ul>
            )}
        </section>
    );
}

interface AlbumNewPopupProps {
    onClose: () => void;
    onSubmit: (name: string, publicState: number) => void;
    error: string | null;
}

function AlbumNewPopup({ onClose, onSubmit, error }: AlbumNewPopupProps) {
    const [name, setName] = useState("");
    const [isPublic, setIsPublic] = useState(false);
    const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!name.trim()) return;
        onSubmit(name, isPublic ? 1 : 0);
    };

    return (
        <div className={styles.popupOverlay}>
            <div className={styles.popup}>
                {error && <p className={styles.error}>{error}</p>}
                <form  onSubmit={handleSubmit}>
                    <div className={styles.fieldRow}>
                    <input
                        type="text"
                        placeholder="Album Name"
                        className={styles.input}
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                    />
                    <button
                        type="button"
                        className={styles.toggleButton}
                        onClick={() => setIsPublic((current) => !current)}
                        aria-pressed={isPublic}
                        aria-label={isPublic ? "Set album to private" : "Set album to public"}
                        title={isPublic ? "Public" : "Private"}
                    >
                        {isPublic ? (
                            <Globe className={`${styles.iconSm} ${styles.publicIcon}`} />
                        ) : (
                            <Lock className={styles.iconSm} />
                        )}
                    </button>
                    </div>
                    <div className={styles.actions}>
                        <button type="submit" className={styles.primaryButton}>
                            Create
                        </button>
                        <button type="button" onClick={onClose} className={styles.secondaryButton}>
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}