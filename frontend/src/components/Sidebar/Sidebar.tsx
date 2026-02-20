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
import { useContext } from "react";
import { UserContext } from "../Auth/AuthGate";

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
    { label: "Tags", href: "/tags", icon: Tags, permissions: ["read:tags"] },
    { label: "Plugins", href: "/plugins", icon: Puzzle },
    {
        label: "Accounts",
        href: "/accounts",
        icon: Users,
        permissions: ["read:user"],
    },
    { label: "Jobs", href: "/jobs", icon: Briefcase },
    { label: "Settings", href: "/settings", icon: Settings },
];

export default function Sidebar() {
    const pathname = usePathname();
    const router = useRouter();

    const user = useContext(UserContext);
    const userPerms = user?.permissions?.map((perm) => perm.permission);

    const onLogout = () => {
        LogoutHandler();
        router.replace("/login");
        router.refresh();
    };

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
                    let hasPermission = true;
                    if (item.permissions) {
                        item.permissions.forEach((perm) => {
                            if (!userPerms?.includes(perm)) {
                                hasPermission = false;
                            }
                        });
                    }

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

                <button
                    type="button"
                    className={`${styles.item} ${styles.logoutItem}`}
                    onClick={onLogout}
                >
                    <span className={styles.itemInner}>
                        <LogOut
                            className={styles.itemIcon}
                            aria-hidden="true"
                        />
                        <span className={styles.itemLabel}>Log out</span>
                    </span>
                </button>
            </nav>
        </div>
    );
}
