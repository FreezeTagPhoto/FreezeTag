import { useContext, useEffect, useState } from "react";
import styles from "./ModifyPerms.module.css";
import { UserContext } from "@/components/Auth/AuthGate";
import {
    ExtractPermsList,
    Perm,
    UserHasPerm,
} from "@/api/permissions/permshelpers";
import PermsGetter from "@/api/permissions/permsgetter";
import PermsLister from "@/api/permissions/permslister";
import PermsAdder from "@/api/permissions/permsadder";
import PermsRevoker from "@/api/permissions/permsrevoker";

export type ModifyPermsProps = {
    onClose: () => void;
    userId: number;
    username: string;
};

export default function ModifyPerms({
    onClose,
    userId,
    username,
}: ModifyPermsProps) {
    const currentUser = useContext(UserContext);
    const currentCanEdit = UserHasPerm(currentUser, "write:permissions");

    const [checkedPerms, setCheckedPerms] = useState<string[] | undefined>(
        undefined,
    );
    const [newCheckedPerms, setNewCheckedPerms] = useState<
        string[] | undefined
    >(undefined);
    const [allPerms, setAllPerms] = useState<Perm[]>([]);

    useEffect(() => {
        (async () => {
            const result = await PermsGetter(userId);
            if (result.ok) {
                const perms = ExtractPermsList({
                    user_id: userId,
                    permissions: result.value,
                });
                setCheckedPerms(perms);
                setNewCheckedPerms(perms);
            } else {
                console.error(
                    `Couldn't get perms! Error ${result.error.status}: ${result.error.message}`,
                );
            }
        })();
        (async () => {
            const result = await PermsLister();
            if (result.ok) {
                setAllPerms(result.value);
            } else {
                console.error(
                    `Couldn't get perms! Error ${result.error.status}: ${result.error.message}`,
                );
            }
        })();
    }, [userId]);

    const onSubmit = async () => {
        if (newCheckedPerms === undefined || checkedPerms === undefined) {
            console.error("One of the checked perms arrays was undefined!");
            return;
        }

        const addedPerms = newCheckedPerms.filter(
            (v) => !checkedPerms.includes(v),
        );
        const removedPerms = checkedPerms.filter(
            (v) => !newCheckedPerms.includes(v),
        );

        if (addedPerms.length > 0) {
            const result = await PermsAdder(userId, addedPerms);
            if (!result.ok) {
                console.error(
                    `Couldn't add perms! Error ${result.error.status}: ${result.error.message}`,
                );
            }
        }

        if (removedPerms.length > 0) {
            const result = await PermsRevoker(userId, removedPerms);
            if (!result.ok) {
                console.error(
                    `Couldn't revoke perms! Error ${result.error.status}: ${result.error.message}`,
                );
            }
        }
        setCheckedPerms(newCheckedPerms);

        onClose();
    };

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>
                        {currentCanEdit ? "Modify" : "View"} Perms for{" "}
                        {username}
                    </h1>
                    <p className={styles.subtle}>Perms Modification</p>
                </header>
                <div className={styles.perms_container}>
                    {checkedPerms !== undefined &&
                        newCheckedPerms !== undefined &&
                        allPerms.map((perm) => (
                            <div
                                key={perm.permission}
                                className={`${styles.perms_row}`}
                            >
                                <input
                                    className={styles.checkbox}
                                    disabled={
                                        !currentCanEdit ||
                                        currentUser?.user_id === userId
                                    }
                                    type="checkbox"
                                    defaultChecked={checkedPerms.includes(
                                        perm.permission,
                                    )}
                                    onChange={(e) => {
                                        if (e.currentTarget.checked) {
                                            setNewCheckedPerms([
                                                ...new Set([
                                                    perm.permission,
                                                    ...newCheckedPerms,
                                                ]),
                                            ]);
                                        } else {
                                            setNewCheckedPerms(
                                                newCheckedPerms.filter(
                                                    (value) =>
                                                        value !==
                                                        perm.permission,
                                                ),
                                            );
                                        }
                                    }}
                                ></input>
                                <div
                                    className={`${styles.perms_item} ${styles.text}`}
                                    title={perm.permission}
                                >
                                    <p className={styles.text_preview}>
                                        {perm.permission}
                                    </p>
                                </div>
                                <div
                                    className={`${styles.perms_item} ${styles.text}`}
                                    title={perm.name}
                                >
                                    <p className={styles.text_preview}>
                                        {perm.name}
                                    </p>
                                </div>
                                <div
                                    className={`${styles.perms_item} ${styles.description}`}
                                    title={perm.description}
                                >
                                    <p className={styles.text_preview}>
                                        {perm.description}
                                    </p>
                                </div>
                            </div>
                        ))}
                </div>
                {currentCanEdit && (
                    <button
                        type="button"
                        className={styles.submit_perms}
                        onClick={() => onSubmit()}
                    >
                        Submit Changes
                    </button>
                )}
            </div>
        </div>
    );
}
