"use client";

import { useCallback, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Check, Plus, Trash2, Globe, Lock, UserIcon, Share2, ChevronDown, Loader2, X } from "lucide-react";
import AlbumImagesGetter from "@/api/albums/albumimagesgetter";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import styles from "./AlbumDetailPage.module.css";
import AlbumGetter from "@/api/albums/albumgetter";
import AlbumRenamer from "@/api/albums/albumrenamer";
import AlbumDeleter from "@/api/albums/albumdeleter";
import ProfilePictureGetter from "@/api/users/profilepicturegetter";
import UserLister, { User } from "@/api/users/userlister";
import { UserContext } from "@/components/Auth/AuthGate";
import { UserHasPerm } from "@/api/permissions/permshelpers";

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

    type PermissionLevel = "owner" | "share" | "none";
    function ShareMenu({
        albumId,
        busy,
    }: {
        albumId: number;
        busy: boolean;
    }) {
        const currentUser = useContext(UserContext);
        const canViewUsers = UserHasPerm(currentUser, "read:user");

        const [isOpen, setIsOpen] = useState(false);
        const [shareableUsers, setShareableUsers] = useState<User[]>([]);
        const [permissions, setPermissions] = useState<
            Map<number, PermissionLevel>
        >(new Map());
        const [shareLoading, setShareLoading] = useState(false);
        const [shareError, setShareError] = useState<string | null>(null);

        useEffect(() => {
            if (!isOpen || !canViewUsers) return;

            let mounted = true;
            const loadData = async () => {
                setShareLoading(true);
                setShareError(null);

                const [usersRes] = await Promise.all([UserLister()]);

                if (mounted) {
                    if (usersRes.ok) {
                        setShareableUsers(usersRes.value);
                    } else {
                        if (
                            usersRes.error.status === 401 ||
                            usersRes.error.status === 403
                        ) {
                            setIsOpen(false);
                            setShareableUsers([]);
                            setShareError(null);
                        } else {
                            setShareError("Failed to load user list.");
                        }
                    }
                    setPermissions(new Map());
                    setShareLoading(false);
                }
            };

            loadData();
            return () => {
                mounted = false;
            };
        }, [isOpen, albumId, canViewUsers]);

        if (!canViewUsers) return null;

        const handlePermissionChange = async (
            userId: number,
            newLevel: PermissionLevel,
        ) => {
            const updatedMap = new Map(permissions);
            if (newLevel === "none") updatedMap.delete(userId);
            else updatedMap.set(userId, newLevel);
            setPermissions(updatedMap);
            // await UpdateAlbumPermission(albumId, userId, newLevel);
        };

        return (
            <div className={styles.shareWrap}>
                <button
                    type="button"
                    className={styles.button}
                    onClick={() => setIsOpen(!isOpen)}
                    disabled={busy}
                >
                    <Share2 size={16} /> Share <ChevronDown size={16} />
                </button>

                {isOpen && (
                    <div className={styles.shareDropdown}>
                        <div className={styles.shareHeaderRow}>
                            <p className={styles.shareTitle}>Share album</p>
                            <button
                                type="button"
                                className={styles.shareCloseButton}
                                onClick={() => setIsOpen(false)}
                            >
                                <X size={16} />
                            </button>
                        </div>

                        {shareError && <p className={styles.error}>{shareError}</p>}

                        {shareLoading ? (
                            <p className={styles.shareLoadingRow}>
                                <Loader2
                                    className={styles.spinningIcon}
                                    size={16}
                                />{" "}
                                Loading...
                            </p>
                        ) : shareableUsers.length === 0 ? (
                            <p className={styles.shareEmptyState}>
                                No users found.
                            </p>
                        ) : (
                            <div className={styles.shareList}>
                                {shareableUsers.map((user) => (
                                    <UserShareRow
                                        key={user.id}
                                        user={user}
                                        currentLevel={
                                            permissions.get(user.id) || "none"
                                        }
                                        onChange={(newLevel) =>
                                            handlePermissionChange(
                                                user.id,
                                                newLevel,
                                            )
                                        }
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                )}
            </div>
        );
    }

    function UserShareRow({
        user,
        currentLevel,
        onChange,
    }: {
        user: User;
        currentLevel: PermissionLevel;
        onChange: (level: PermissionLevel) => void;
    }) {
        return (
            <div className={styles.shareUserRow}>
                <UserAvatar userId={user.id} username={user.username} />

                <div className={styles.shareUserMeta}>
                    <span className={styles.shareUserName}>{user.username}</span>
                </div>

                <div className={styles.sharePermissionControl}>
                    <select
                        className={styles.shareRoleSelect}
                        value={currentLevel}
                        onChange={(e) =>
                            onChange(e.target.value as PermissionLevel)
                        }
                        disabled={currentLevel === "owner"}
                    >
                        {currentLevel === "owner" && (
                            <option value="owner">Owner</option>
                        )}
                        <option value="owner">Owner</option>
                        <option value="viewer">Viewer</option>
                        <option value="none">No Access</option>
                    </select>
                </div>
            </div>
        );
    }

    function UserAvatar({
        userId,
        username,
    }: {
        userId: number;
        username: string;
    }) {
        const [pfpUrl, setPfpUrl] = useState<string | null>(null);

        useEffect(() => {
            let mounted = true;
            let objectUrl: string | null = null;

            const loadPfp = async () => {
                const result = await ProfilePictureGetter(userId);

                if (mounted && result.ok) {
                    objectUrl = URL.createObjectURL(result.value);
                    setPfpUrl(objectUrl);
                }
            };
            loadPfp();
            return () => {
                mounted = false;
                if (objectUrl) {
                    URL.revokeObjectURL(objectUrl);
                }
            };
        }, [userId]);

        if (pfpUrl) {
            return (
                <img src={pfpUrl} alt={username} className={styles.shareAvatar} />
            );
        }

        return (
            <div className={styles.shareAvatarFallback}>
                {username ? username.charAt(0) : <UserIcon size={20} />}
            </div>
        );
    }
}
