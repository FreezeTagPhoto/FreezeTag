import { useState } from "react";
import styles from "./DeleteUser.module.css";
import UserDeleter from "@/api/users/userdeleter";

export type DeleteUserProps = {
    onClose: () => void;
    userId: number;
    username: string;
};

export default function DeleteUser({
    onClose,
    userId,
    username,
}: DeleteUserProps) {
    const [error, setError] = useState("");

    async function onDelete() {
        const result = await UserDeleter(userId);

        if (result.ok) {
            onClose();
        } else {
            setError(`Error ${result.error.status}: ${result.error.message}`);
        }
    }

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Are you sure?</h1>
                    <p className={styles.subtle}>Deleting {username}</p>
                </header>
                {error && (
                    <p className={`${styles.calloutError} ${styles.callout}`}>
                        {error}
                    </p>
                )}
                <div className={styles.buttons}>
                    <button
                        className={`${styles.primary} ${styles.yes_button}`}
                        onClick={onDelete}
                    >
                        Yes, Delete User {userId}
                    </button>
                    <button className={styles.primary} onClick={onClose}>
                        No, Keep User {userId}
                    </button>
                </div>
            </div>
        </div>
    );
}
