"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import logoUrl from "@/icons/freezetag+text.svg";
import cubeLogoUrl from "@/icons/freezetag.svg";
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
    PanelLeftClose,
    PanelLeftOpen,
    ChevronsLeft,
} from "lucide-react";

import { useContext, useEffect, useMemo, useState } from "react";
import { UserContext } from "../Auth/AuthGate";
import { ProfilePictureContext } from "../Auth/ProfilePictureContext";
import { ExtractPermsList } from "@/api/permissions/permshelpers";
import UserGetter from "@/api/users/usergetter";
import ProfilePictureGetter from "@/api/users/profilepicturegetter";

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
    profilePictureUrl,
    collapsed,
}: {
    username: string;
    userId: number;
    profilePictureUrl: string | null;
    collapsed: boolean;
}) {
    const initial =
        username && username.trim().length > 0
            ? username.trim()[0]!.toUpperCase()
            : (userId.toString()[0] ?? "?");

    return (
        <div
            className={`${styles.accountInfo} ${collapsed ? styles.accountInfoCollapsed : ""}`}
            aria-label="Signed in user"
            title={collapsed ? username : undefined}
        >
            <span className={styles.itemInner}>
                {profilePictureUrl ? (
                    <span
                        className={`${styles.itemIcon} ${styles.avatarIcon}`}
                        aria-hidden="true"
                        style={{
                            position: "relative",
                            overflow: "hidden",
                            borderRadius: 9999,
                            border: "var(--border-info)",
                            background: "var(--mantle)",
                        }}
                    >
                        <img
                            src={profilePictureUrl}
                            alt=""
                            style={{
                                width: "100%",
                                height: "100%",
                                objectFit: "cover",
                                display: "block",
                            }}
                        />
                    </span>
                ) : (
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
                )}

                {!collapsed && (
                    <span
                        className={styles.itemLabel}
                        style={{ lineHeight: 1.15 }}
                    >
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
                )}
            </span>
        </div>
    );
}

type SidebarProps = {
    collapsed?: boolean;
    onToggleCollapsed?: () => void;
    /** Called when the sidebar should close (mobile overlay or after nav). */
    onMobileClose?: () => void;
};

export default function Sidebar({
    collapsed = false,
    onToggleCollapsed,
    onMobileClose,
}: SidebarProps) {
    const pathname = usePathname();
    const router = useRouter();

    const user = useContext(UserContext);
    const { profilePictureVersion } = useContext(ProfilePictureContext);
    const userPerms = useMemo(() => ExtractPermsList(user) ?? [], [user]);

    const [username, setUsername] = useState<string>("");
    const [profilePictureUrl, setProfilePictureUrl] = useState<string | null>(
        null,
    );

    useEffect(() => {
        if (!user) return;

        (async () => {
            const result = await UserGetter(user.user_id);
            if (!result.ok) {
                console.error(`User Getter Error! ${result.error.message}`);
                return;
            }
            setUsername(result.value.username);
        })();
    }, [user]);

    useEffect(() => {
        if (!user) return;

        let alive = true;
        let urlToRevoke: string | null = null;

        (async () => {
            const result = await ProfilePictureGetter(user.user_id);

            if (!alive) return;

            if (!result.ok) {
                console.warn(
                    `Profile picture fetch failed (${result.error.status}): ${result.error.message}`,
                );
                setProfilePictureUrl(null);
                return;
            }

            urlToRevoke = URL.createObjectURL(result.value);
            setProfilePictureUrl(urlToRevoke);
        })();

        return () => {
            alive = false;
            if (urlToRevoke) URL.revokeObjectURL(urlToRevoke);
        };
    }, [user, profilePictureVersion]);

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
        <div
            className={`${styles.wrap} ${collapsed ? styles.wrapCollapsed : ""}`}
        >
            {/* Brand / logo – no buttons here */}
            <div className={styles.brand}>
                <Link
                    href="/"
                    className={`${styles.logoWrap} ${collapsed ? styles.logoWrapCollapsed : ""}`}
                    title="FreezeTag"
                    onClick={() => onMobileClose?.()}
                >
                    <Image
                        src={collapsed ? cubeLogoUrl : logoUrl}
                        alt="FreezeTag"
                        fill
                        priority
                        sizes={collapsed ? "48px" : "240px"}
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
                        return <div key={item.label} className={styles.noop} />;
                    }

                    return (
                        <Link
                            key={item.label}
                            href={item.href}
                            className={`${styles.item} ${
                                isActive ? styles.itemActive : ""
                            } ${collapsed ? styles.itemCollapsed : ""}`}
                            aria-current={isActive ? "page" : undefined}
                            title={collapsed ? item.label : undefined}
                            onClick={() => onMobileClose?.()}
                        >
                            <span className={styles.itemInner}>
                                <Icon
                                    className={styles.itemIcon}
                                    aria-hidden="true"
                                />
                                {!collapsed && (
                                    <span className={styles.itemLabel}>
                                        {item.label}
                                    </span>
                                )}
                            </span>
                        </Link>
                    );
                })}

                <div className={styles.sectionBreak} aria-hidden="true">
                    <div className={styles.sectionLine} />
                </div>

                <div className={styles.collapseSidebar}>
                    {/* Desktop collapse/expand */}
                    {onToggleCollapsed && (
                        <button
                            type="button"
                            className={`${styles.item} ${styles.collapseToggle} ${
                                collapsed ? styles.itemCollapsed : ""
                            } ${styles.collapseToggleDesktop}`}
                            onClick={onToggleCollapsed}
                            aria-label={
                                collapsed
                                    ? "Expand sidebar"
                                    : "Collapse sidebar"
                            }
                            title={
                                collapsed
                                    ? "Expand sidebar"
                                    : "Collapse sidebar"
                            }
                        >
                            <span className={styles.itemInner}>
                                {collapsed ? (
                                    <PanelLeftOpen
                                        className={styles.itemIcon}
                                        aria-hidden="true"
                                    />
                                ) : (
                                    <PanelLeftClose
                                        className={styles.itemIcon}
                                        aria-hidden="true"
                                    />
                                )}
                                {!collapsed && (
                                    <span className={styles.itemLabel}>
                                        Collapse Sidebar
                                    </span>
                                )}
                            </span>
                        </button>
                    )}

                    {/* Mobile close */}
                    <button
                        type="button"
                        className={`${styles.item} ${styles.collapseToggle} ${styles.collapseToggleMobile}`}
                        onClick={onMobileClose}
                        aria-label="Close navigation"
                    >
                        <span className={styles.itemInner}>
                            <ChevronsLeft
                                className={styles.itemIcon}
                                aria-hidden="true"
                            />
                            <span className={styles.itemLabel}>Close</span>
                        </span>
                    </button>
                </div>

                <div className={styles.bottomDock}>
                    <AccountInfo
                        username={username}
                        userId={user.user_id}
                        profilePictureUrl={profilePictureUrl}
                        collapsed={collapsed}
                    />

                    <div
                        role="button"
                        tabIndex={0}
                        className={`${styles.item} ${styles.logoutItem} ${
                            collapsed ? styles.itemCollapsed : ""
                        }`}
                        onClick={onLogout}
                        onKeyDown={onLogoutKeyDown}
                        aria-label="Log out"
                        title={collapsed ? "Log out" : undefined}
                    >
                        <span className={styles.itemInner}>
                            <LogOut
                                className={styles.itemIcon}
                                aria-hidden="true"
                            />
                            {!collapsed && (
                                <span className={styles.itemLabel}>
                                    Log out
                                </span>
                            )}
                        </span>
                    </div>
                </div>
            </nav>
        </div>
    );
}
