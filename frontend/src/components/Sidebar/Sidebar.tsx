"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import logoUrl from "@/icons/freezetag+text.svg";
import styles from "./Sidebar.module.css";
import LogoutHandler from "@/api/auth/logouthandler";

import type { LucideIcon } from "lucide-react";
import {
    Images,
    Upload,
    Tags,
    Settings,
    Briefcase,
    Puzzle,
    Users,
    LogOut,
} from "lucide-react";

import { useContext, useEffect, useMemo, useState } from "react";
import { UserContext } from "../Auth/AuthGate";
import { ExtractPermsList } from "@/api/permissions/permshelpers";
import UserGetter from "@/api/users/usergetter";

type NavItem = {
    label: string;
    href: string;
    icon: LucideIcon;
    permissions?: string[];
};

const navItems: NavItem[] = [
    { label: "Gallery", href: "/", icon: Images },
    {
        label: "Upload",
        href: "/upload",
        icon: Upload,
        permissions: ["write:files"],
    },
    { label: "Jobs", href: "/jobs", icon: Briefcase },
    {
        label: "Tags",
        href: "/tags",
        icon: Tags,
        permissions: ["read:tags"],
    },
    { label: "Plugins", href: "/plugins", icon: Puzzle },
    {
        label: "Accounts",
        href: "/accounts",
        icon: Users,
        permissions: ["read:user"],
    },
    { label: "Settings", href: "/settings", icon: Settings },
];

function AccountInfo({
    username,
    userId,
}: {
    username: string;
    userId: number;
}) {
    const initial =
        username && username.trim().length > 0
            ? username.trim()[0]!.toUpperCase()
            : (userId.toString()[0] ?? "?");

    return (
        <div className={styles.accountInfo} aria-label="Signed in user">
            <span className={styles.itemInner}>
                {/* Avatar placeholder */}
                <span
                    className={`${styles.itemIcon} ${styles.avatarIcon}`}
                    aria-hidden="true"
                    style={{
                        display: "grid",
                        placeItems: "center",
                        borderRadius: 9999,
                        border: "var(--border-info)",
                        background: "var(--mantle)",
                        color: "var(--text)",
                        fontWeight: 800,
                        lineHeight: 1,
                        userSelect: "none",
                    }}
                >
                    {initial}
                </span>

                <span className={styles.itemLabel} style={{ lineHeight: 1.15 }}>
                    <span
                        style={{
                            display: "block",
                            fontWeight: 700,
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                            maxWidth: "100%",
                        }}
                        title={username}
                    >
                        {username}
                    </span>
                    <span
                        style={{
                            display: "block",
                            opacity: 0.8,
                            fontSize: "0.85em",
                        }}
                    >
                        ID {userId}
                    </span>
                </span>
            </span>
        </div>
    );
}

export default function Sidebar() {
    const pathname = usePathname();
    const router = useRouter();

    const user = useContext(UserContext);
    const userPerms = useMemo(() => ExtractPermsList(user) ?? [], [user]);

    const [username, setUsername] = useState<string>("");

    useEffect(() => {
        if (!user) return;

        (async () => {
            const result = await UserGetter(user.user_id);
            if (!result.ok) {
                console.error(`User Lister Error! ${result.error.message}`);
                return;
            }
            setUsername(result.value.username);
        })();
    }, [user]);

    const onLogout = async () => {
        await LogoutHandler();
        router.replace("/login");
        router.refresh();
    };

    const onLogoutKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
        if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onLogout();
        }
    };

    if (!user) return null;

    return (
        <div className={styles.wrap}>
            <div className={styles.brand}>
                <Link href="/" className={styles.logoWrap}>
                    <Image
                        src={logoUrl}
                        alt="FreezeTag"
                        fill
                        priority
                        sizes="240px"
                        className={styles.logoImg}
                    />
                </Link>
            </div>

            <nav className={styles.menu}>
                {navItems.map((item) => {
                    const isActive =
                        item.href === "/"
                            ? pathname === "/"
                            : pathname === item.href ||
                              pathname.startsWith(item.href + "/");

                    const Icon = item.icon;

                    const hasPermission = item.permissions
                        ? item.permissions.every((perm) =>
                              userPerms.includes(perm),
                          )
                        : true;

                    if (!hasPermission) {
                        return (
                            <div key={item.label} className={styles.noop}></div>
                        );
                    }

                    return (
                        <Link
                            key={item.label}
                            href={item.href}
                            className={`${styles.item} ${
                                isActive ? styles.itemActive : ""
                            }`}
                            aria-current={isActive ? "page" : undefined}
                        >
                            <span className={styles.itemInner}>
                                <Icon
                                    className={styles.itemIcon}
                                    aria-hidden="true"
                                />
                                <span className={styles.itemLabel}>
                                    {item.label}
                                </span>
                            </span>
                        </Link>
                    );
                })}

                <div className={styles.sectionBreak} aria-hidden="true">
                    <div className={styles.sectionLine} />
                </div>

                <div className={styles.bottomDock}>
                    <AccountInfo username={username} userId={user.user_id} />
                    <div
                        role="button"
                        tabIndex={0}
                        className={`${styles.item} ${styles.logoutItem}`}
                        onClick={onLogout}
                        onKeyDown={onLogoutKeyDown}
                        aria-label="Log out"
                    >
                        <span className={styles.itemInner}>
                            <LogOut
                                className={styles.itemIcon}
                                aria-hidden="true"
                            />
                            <span className={styles.itemLabel}>Log out</span>
                        </span>
                    </div>
                </div>
            </nav>
        </div>
    );
}
