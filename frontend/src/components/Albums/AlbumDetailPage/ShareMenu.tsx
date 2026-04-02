import { useState, useEffect } from "react";
import { Share2, ChevronDown, Loader2, X, User as UserIcon } from "lucide-react";
import UserLister, { User } from "@/api/users/userlister"; 
import styles from "./AlbumDetailPage.module.css";
import ProfilePictureGetter from "@/api/users/profilepicturegetter";

export type PermissionLevel = "owner" | "share" | "none";

export default function ShareMenu({ albumId, busy }: { albumId: number; busy: boolean }) {
    const [isOpen, setIsOpen] = useState(false);
    const [shareableUsers, setShareableUsers] = useState<User[]>([]);
    const [permissions, setPermissions] = useState<Map<number, PermissionLevel>>(new Map());
    const [shareLoading, setShareLoading] = useState(false);
    const [shareError, setShareError] = useState<string | null>(null);

    useEffect(() => {
        if (!isOpen) return;

        let mounted = true;
        const loadData = async () => {
            setShareLoading(true);
            setShareError(null);
            
            const [usersRes] = await Promise.all([ 
                UserLister() 
            ]);
            
            if (mounted) {
                if (usersRes.ok) {
                    setShareableUsers(usersRes.value);
                } else {
                    setShareError("Failed to load user list.");
                }
                setPermissions(new Map()); 
                setShareLoading(false);
            }
        };

        loadData();
        return () => { mounted = false; };
    }, [isOpen, albumId]);

    const handlePermissionChange = async (userId: number, newLevel: PermissionLevel) => {
        const updatedMap = new Map(permissions);
        if (newLevel === "none") updatedMap.delete(userId);
        else updatedMap.set(userId, newLevel);
        setPermissions(updatedMap);
        // await UpdateAlbumPermission(albumId, userId, newLevel);
    };

    return (
        <div className={styles.shareWrap}>
            <button type="button" className={styles.button} onClick={() => setIsOpen(!isOpen)} disabled={busy}>
                <Share2 size={16} /> Share <ChevronDown size={16} />
            </button>

            {isOpen && (
                <div className={styles.shareDropdown}>
                    <div className={styles.shareHeaderRow}>
                        <p className={styles.shareTitle}>Share album</p>
                        <button type="button" className={styles.shareCloseButton} onClick={() => setIsOpen(false)}>
                            <X size={16} />
                        </button>
                    </div>

                    {shareError && <p className={styles.error}>{shareError}</p>}

                    {shareLoading ? (
                        <p className={styles.shareLoadingRow}>
                            <Loader2 className={styles.spinningIcon} size={16} /> Loading...
                        </p>
                    ) : shareableUsers.length === 0 ? (
                        <p className={styles.shareEmptyState}>No users found.</p>
                    ) : (
                        <div className={styles.shareList}>
                            {shareableUsers.map(user => (
                                <UserShareRow 
                                    key={user.id} 
                                    user={user} 
                                    currentLevel={permissions.get(user.id) || "none"} 
                                    onChange={(newLevel) => handlePermissionChange(user.id, newLevel)}
                                />
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

function UserShareRow({ user, currentLevel, onChange }: { user: User; currentLevel: PermissionLevel; onChange: (level: PermissionLevel) => void; }) {
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
                    onChange={(e) => onChange(e.target.value as PermissionLevel)}
                    disabled={currentLevel === "owner"} 
                >
                    {currentLevel === "owner" && <option value="owner">Owner</option>}
                    <option value="owner">Owner</option>
                    <option value="viewer">Viewer</option>
                    <option value="none">No Access</option>
                </select>
            </div>
        </div>
    );
}

function UserAvatar({ userId, username }: { userId: number, username: string }) {
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
        return <img src={pfpUrl} alt={username} className={styles.shareAvatar} />;
    }

    return (
        <div className={styles.shareAvatarFallback}>
            {username ? username.charAt(0) : <UserIcon size={20} />}
        </div>
    );
}