"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, Plus, Trash2 } from "lucide-react";
import AlbumImagesGetter from "@/api/albums/albumimagesgetter";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import styles from "./AlbumDetailPage.module.css";
import AlbumGetter from "@/api/albums/albumgetter";
import AlbumRenamer from "@/api/albums/albumrenamer";
import ShareMenu from "./ShareMenu";
import AlbumDeleter from "@/api/albums/albumdeleter";

export default function AlbumDetailPage({ albumId }: { albumId: number }) {
    const [name, setName] = useState<string>("");
    const [imageIds, setImageIds] = useState<number[]>([]);
    const [loading, setLoading] = useState(true);
    const [busy, setBusy] = useState(false);

    const canShowShare = true;
    const canManageSharing = true;

    const loadData = useCallback(async () => {
        setLoading(true);

        const [albumRes, imagesRes] = await Promise.all([
            AlbumGetter(albumId),
            AlbumImagesGetter(albumId),
        ]);
        if (albumRes.ok) setName(albumRes.value.name);
        if (imagesRes.ok) setImageIds(imagesRes.value);

        setLoading(false);
    }, [albumId]);

    const handleRename = async (newName: string) => {
        if (!newName.trim() || newName === name) return;
        setBusy(true);
        const result = await AlbumRenamer(albumId, newName);
        if (result.ok) {
            setName(newName);
        } else {
            console.error("Failed to rename album:", result.error);
        }
        setBusy(false);
    };

    useEffect(() => {
        loadData();
    }, [loadData]);

    const handleDelete = async () => {
        const result = window.confirm(
            "Delete Album? This action cannot be undone.",
        );
        if (!result) return;
        const deleteResult = await AlbumDeleter(albumId);
        if (deleteResult.ok) {
            window.location.href = "/albums";
        } else {
            window.alert("Failed to delete album. Please try again.");
        }
    };

    return (
        <section className={styles.wrap}>
            <nav>
                <Link href="/albums" className={styles.backLink}>
                    <ArrowLeft size={16} /> Back to albums
                </Link>
            </nav>

            <header className={styles.header}>
                <div className={styles.titleBlock}>
                    <RenameableTitle
                        key={name}
                        name={name}
                        onRename={handleRename}
                        busy={busy}
                    />
                    {!loading && (
                        <p className={styles.meta}>{imageIds.length} images</p>
                    )}
                </div>

                <div className={styles.settings}>
                    {canShowShare && (
                        <ShareMenu busy={busy} albumId={albumId} />
                    )}

                    <button
                        className={`${styles.button} ${styles.buttonDanger}`}
                        onClick={handleDelete}
                        disabled={busy}
                    >
                        <Trash2 size={16} /> Delete
                    </button>

                    {canManageSharing && (
                        <Link
                            href={`/albums/${encodeURIComponent(albumId.toString())}/edit`}
                            className={styles.button}
                        >
                            <Plus size={16} /> Manage Images
                        </Link>
                    )}
                </div>
            </header>

            {loading ? <p>Loading...</p> : <MainGallery image_ids={imageIds} />}
        </section>
    );
}

function RenameableTitle({
    name,
    onRename,
    busy,
}: {
    name: string;
    onRename: (n: string) => Promise<void> | void;
    busy: boolean;
}) {
    const [rename, setRename] = useState(false);
    const [editName, setEditName] = useState(name);

    if (!rename) {
        return (
            <button
                className={styles.titleField}
                onClick={() => setRename(true)}
            >
                <h1 className={styles.title}>{name}</h1>
            </button>
        );
    }

    return (
        <div className={styles.renameRow}>
            <input
                autoFocus
                className={styles.titleField}
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && onRename(editName)}
                onBlur={() => {
                    if (!busy) {
                        setRename(false);
                        setEditName(name);
                    }
                }}
                disabled={busy}
            />
            <button
                className={styles.confirmButton}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => onRename(editName)}
                disabled={busy || !editName.trim() || editName === name}
            >
                <Check size={28} />
            </button>
        </div>
    );
}
