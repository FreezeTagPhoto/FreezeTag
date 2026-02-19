"use client";
import { useContext, useEffect, useState } from "react";
import styles from "./page.module.css";
import UserLister, { User } from "@/api/users/userlister";
import { formatDate } from "@/common/gallery/format";
import { UserContext } from "@/components/Auth/AuthGate";
import CreateUser from "@/components/ManageUsers/CreateUser";
import { SquarePen, Trash, UserPlus } from "lucide-react";

export default function Home() {
    const [users, setUsers] = useState<User[]>([]);
    const [userDigits, setUserDigits] = useState<number>(0);

    const [creatingUser, setCreatingUser] = useState<boolean>(false);

    const user = useContext(UserContext);
    const isCurrent = (id: number) => id === user?.user_id;

    useEffect(() => {
        (async () => {
            const result = await UserLister();
            if (result.ok) {
                setUsers(result.value);
                setUserDigits(
                    Math.max(...result.value.map((user) => user.id)).toString()
                        .length,
                );
            } else {
                console.error(`User Lister Error! ${result.error.message}`);
            }
        })();
    }, [creatingUser]);

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Account Management</h1>
            <div className={styles.user_container}>
                {users.map((user) => (
                    <div
                        key={user.id}
                        className={`${styles.account_row} ${isCurrent(user.id) ? styles.account_row_current : ""}`}
                    >
                        <p className={`${styles.account_item} `}>
                            {user.id.toString().padStart(userDigits, "0")}
                        </p>
                        <div
                            className={`${styles.account_item} ${styles.username}`}
                            title={user.username}
                        >
                            <p className={styles.username_preview}>
                                {user.username}
                            </p>
                        </div>
                        <p className={`${styles.account_item} ${styles.date}`}>
                            {formatDate(user.created_at)}
                        </p>
                        <p className={styles.account_item}>
                            <SquarePen className={styles.icon} />
                            View/Modify Perms
                        </p>
                        <p className={styles.account_item}>
                            <Trash className={styles.icon} />
                            Delete User
                        </p>
                    </div>
                ))}
            </div>
            <button
                type="button"
                className={styles.create_user}
                disabled={
                    !(
                        user?.permissions &&
                        user?.permissions
                            .map((perm) => perm.permission)
                            .includes("create:user")
                    )
                }
                onClick={() => setCreatingUser(true)}
            >
                <UserPlus
                    className={styles.create_user_icon}
                    aria-hidden={true}
                ></UserPlus>
                Create User
            </button>
            {creatingUser && (
                <CreateUser onClose={() => setCreatingUser(false)}></CreateUser>
            )}
        </main>
    );
}
