"use client";
import { useContext, useEffect, useState } from "react";
import styles from "./page.module.css";
import UserLister, { User } from "@/api/users/userlister";
import { UserContext } from "@/components/Auth/AuthGate";
import CreateUser from "@/components/ManageUsers/CreateUser/CreateUser";
import { CakeSlice, SquarePen, Trash, UserPlus } from "lucide-react";
import DeleteUser from "@/components/ManageUsers/DeleteUser/DeleteUser";
import ModifyPerms from "@/components/ManageUsers/ModifyPerms/ModifyPerms";

function formatShortDate(
    ts: number | null,
    opts?: { timeZone?: string },
): string {
    if (ts === null) return "—";
    const d = new Date(ts > 1e12 ? ts : ts * 1000);
    return new Intl.DateTimeFormat(undefined, {
        timeZone: opts?.timeZone,
        year: "numeric",
        month: "short",
        day: "numeric",
    }).format(d);
}

function formatLongDate(
    ts: number | null,
    opts?: { timeZone?: string },
): string {
    if (ts === null) return "—";
    const d = new Date(ts > 1e12 ? ts : ts * 1000);
    return new Intl.DateTimeFormat(undefined, {
        timeZone: opts?.timeZone,
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
    }).format(d);
}

export default function Home() {
    const [users, setUsers] = useState<User[]>([]);
    const [userDigits, setUserDigits] = useState<number>(0);

    const [creatingUser, setCreatingUser] = useState<boolean>(false);
    const [modifyingPerms, setModifyingPerms] = useState<number>(-1);
    const [deletingUser, setDeletingUser] = useState<number>(-1);
    const [username, setUsername] = useState<string>("");

    const currentUser = useContext(UserContext);
    const isCurrent = (id: number) => id === currentUser?.user_id;

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
    }, [creatingUser, modifyingPerms, deletingUser]);

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
                        <p
                            className={`${styles.account_item} ${styles.date}`}
                            title={formatLongDate(user.created_at)}
                        >
                            <CakeSlice className={styles.icon} />
                            {formatShortDate(user.created_at)}
                        </p>
                        <button
                            type="button"
                            className={`${styles.account_item} ${styles.account_item_button}`}
                            disabled={
                                !(
                                    currentUser?.permissions &&
                                    currentUser?.permissions
                                        .map((perm) => perm.permission)
                                        .includes("read:permissions")
                                )
                            }
                            onClick={() => setModifyingPerms(user.id)}
                        >
                            <SquarePen className={styles.icon} />
                            <p className={styles.account_item_label}>Perms</p>
                        </button>
                        <button
                            type="button"
                            className={`${styles.account_item} ${styles.account_item_button} ${styles.account_item_delete}`}
                            disabled={
                                !(
                                    currentUser?.permissions &&
                                    currentUser?.permissions
                                        .map((perm) => perm.permission)
                                        .includes("delete:user")
                                ) || isCurrent(user.id)
                            }
                            onClick={() => {
                                setDeletingUser(user.id);
                                setUsername(user.username);
                            }}
                        >
                            <Trash className={styles.icon} />
                            <p className={styles.account_item_label}>Delete</p>
                        </button>
                    </div>
                ))}
            </div>
            <button
                type="button"
                className={styles.create_user}
                disabled={
                    !(
                        currentUser?.permissions &&
                        currentUser?.permissions
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
                <CreateUser onClose={() => setCreatingUser(false)} />
            )}
            {deletingUser !== -1 && (
                <DeleteUser
                    onClose={() => {
                        setDeletingUser(-1);
                        setUsername("");
                    }}
                    userId={deletingUser}
                    username={username}
                />
            )}
            {modifyingPerms !== -1 && (
                <ModifyPerms
                    onClose={() => setModifyingPerms(-1)}
                    userId={modifyingPerms}
                />
            )}
        </main>
    );
}
