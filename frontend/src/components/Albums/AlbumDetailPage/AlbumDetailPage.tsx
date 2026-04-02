"use client";

import { useCallback, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, ChevronDown, Loader2, Plus, Share2, Trash2, X } from "lucide-react";
import AlbumImagesGetter from "@/api/albums/albumimagesgetter";
import { UserContext } from "@/components/Auth/AuthGate";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import styles from "./AlbumDetailPage.module.css";
import AlbumGetter from "@/api/albums/albumgetter";
import AlbumRenamer from "@/api/albums/albumrenamer";
import ShareMenu from "./ShareMenu";

// import your other dependencies...

export default function AlbumDetailPage({ albumId }: { albumId: number }) {
    const [name, setName] = useState<string>("");
    const [id, setId] = useState<number>(albumId);
    const [imageIds, setImageIds] = useState<number[]>([]);
    const [loading, setLoading] = useState(true);
    const [busy, setBusy] = useState(false);

    const canShowShare = true;
    const canManageSharing = true;

    const loadData = useCallback(async () => {
        setLoading(true);

        const [albumRes, imagesRes] = await Promise.all([
            AlbumGetter(albumId),
            AlbumImagesGetter(albumId)
        ]);
        console.log("Full Album Response:", albumRes);
        if (albumRes.ok) setName(albumRes.value.name);
        if (imagesRes.ok) setImageIds(imagesRes.value);

        setLoading(false);
    }, [albumId]);

    const handleRename = async (newName: string) => {
        if (!newName.trim() || newName === name) return;
        setBusy(true);
        const result = await AlbumRenamer(albumId, newName); // implement this API call
        if (result.ok) {
            setName(newName);
        } else {
            // handle error, e.g. show a toast
        }
        setBusy(false);
    }

    useEffect(() => { loadData(); }, [loadData]);

    const handleDelete = () => { /* implement */ };

    return (
        <section className={styles.wrap}>
            <nav>
                <Link href="/albums" className={styles.backLink}>
                    <ArrowLeft size={16} /> Back to albums
                </Link>
            </nav>

            <header className={styles.header}>
                <div className={styles.titleBlock}>
                    <RenameableTitle key={name} name={name} onRename={handleRename} busy={busy} />
                    {!loading && <p className={styles.meta}>{imageIds.length} images</p>}
                </div>

                <div className={styles.settings}>
                    {canShowShare && <ShareMenu busy={busy} albumId={albumId} />}

                    <button className={`${styles.button} ${styles.buttonDanger}`} onClick={handleDelete} disabled={busy}>
                        <Trash2 size={16} /> Delete
                    </button>

                    {canManageSharing && (
                        <Link href={`/albums/${encodeURIComponent(name)}/edit`} className={styles.button}>
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
    onRename, // Change this from setName
    busy 
}: { 
    name: string, 
    onRename: (n: string) => Promise<void> | void, 
    busy: boolean 
}) {
    const [rename, setRename] = useState(false);
    const [editName, setEditName] = useState(name);

    if (!rename) {
        return (
            <button className={styles.titleField} onClick={() => setRename(true)}>
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
                onBlur={() => { if (!busy) { setRename(false); setEditName(name); } }}
                disabled={busy}
            />
            <button
                className={styles.confirmButton}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => onRename(editName)}
                disabled={busy || !editName.trim() || editName === name}>
                <Check size={28} />
            </button>
        </div>
    );
}