import styles from "./ModifyPerms.module.css";

export type CreateUserProps = {
    onClose: () => void;
    userId: number;
};

export default function ModifyPerms({ onClose, userId }: CreateUserProps) {
    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Modify Perms</h1>
                    <p className={styles.subtle}>Perms Modification</p>
                </header>
                {userId}
            </div>
        </div>
    );
}
