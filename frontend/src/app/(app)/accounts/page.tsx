"use client";
import { useEffect, useState } from "react";
import styles from "./page.module.css";
import UserLister, { User } from "@/api/users/userlister";
import { formatDate } from "@/common/gallery/format";

export default function Home() {
    const [users, setUsers] = useState<User[]>([]);

    useEffect(() => {
        (async () => {
            const result = await UserLister();
            if (result.ok) {
                setUsers(result.value);
            } else {
                console.error(`User Lister Error! ${result.error.message}`);
            }
        })();
    }, []);

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Account Management</h1>
            <div className={styles.user_container}>
                {users.map((user) => (
                    <div key={user.id} className={styles.account_row}>
                        <p className={styles.account_item}>{user.id}</p>
                        <p
                            className={`${styles.account_item} ${styles.username}`}
                        >
                            {user.username}
                        </p>
                        <p className={styles.account_item}>
                            {formatDate(user.created_at)}
                        </p>
                        <p className={styles.account_item}>View/Modify Perms</p>
                        <p className={styles.account_item}>Delete User</p>
                    </div>
                ))}
            </div>
            <p className={styles.create_user}>Create User</p>
        </main>
    );
}
