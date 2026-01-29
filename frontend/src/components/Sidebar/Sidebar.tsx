"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import logoUrl from "@/icons/freezetag+text.svg";
import styles from "./Sidebar.module.css";
import { ClearToken } from "@/api/auth/tokenhelpers";

const navItems = [
    { label: "Gallery", href: "/" },
    { label: "Upload", href: "/upload" },
    { label: "Manage", href: "/manage" },
    { label: "Settings", href: "/settings" },
    { label: "Plugins", href: "/plugins" },
    { label: "Accounts", href: "/accounts" },
];

export default function Sidebar() {
    const pathname = usePathname();
    const router = useRouter();

    const onLogout = () => {
        ClearToken();
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

                    return (
                        <Link
                            key={item.label}
                            href={item.href}
                            className={`${styles.item} ${isActive ? styles.itemActive : ""}`}
                            aria-current={isActive ? "page" : undefined}
                        >
                            <span>{item.label}</span>
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
                    <span>Log out</span>
                </button>
            </nav>
        </div>
    );
}
