"use client";

import { useEffect, useState } from "react";
import { Menu } from "lucide-react";
import styles from "@/app/page.module.css";
import Sidebar from "@/components/Sidebar/Sidebar";
import AuthGate from "@/components/Auth/AuthGate";

const STORAGE_KEY = "freezetag:sidebar-collapsed";
const MOBILE_BREAKPOINT = 768;

export default function SidebarShell({
    children,
}: {
    children: React.ReactNode;
}) {
    const [collapsed, setCollapsed] = useState(false);
    const [mobileOpen, setMobileOpen] = useState(false);
    const [isMobile, setIsMobile] = useState(false);

    // Track mobile breakpoint
    useEffect(() => {
        const mq = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT}px)`);
        setIsMobile(mq.matches);
        const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
        mq.addEventListener("change", handler);
        return () => mq.removeEventListener("change", handler);
    }, []);

    // Restore collapsed preference from localStorage
    useEffect(() => {
        try {
            const stored = localStorage.getItem(STORAGE_KEY);
            if (stored === "true") setCollapsed(true);
        } catch {
            // ignore
        }
    }, []);

    const toggleCollapsed = () => {
        setCollapsed((prev) => {
            const next = !prev;
            try {
                localStorage.setItem(STORAGE_KEY, String(next));
            } catch {
                // ignore
            }
            return next;
        });
    };

    const closeMobile = () => setMobileOpen(false);

    return (
        <AuthGate>
            <div className={styles.shell}>
                {mobileOpen && (
                    <div
                        className={styles.mobileBackdrop}
                        onClick={closeMobile}
                        aria-hidden="true"
                    />
                )}

                <aside
                    className={[
                        styles.nav,
                        collapsed && !isMobile ? styles.navCollapsed : "",
                        mobileOpen ? styles.navMobileOpen : "",
                    ]
                        .filter(Boolean)
                        .join(" ")}
                >
                    <Sidebar
                        collapsed={collapsed && !isMobile}
                        onToggleCollapsed={toggleCollapsed}
                        onMobileClose={closeMobile}
                    />
                </aside>

                <div className={styles.content}>
                    <button
                        className={styles.mobileMenuBtn}
                        onClick={() => setMobileOpen(true)}
                        aria-label="Open navigation"
                        type="button"
                    >
                        <Menu size={20} />
                    </button>

                    {children}
                </div>
            </div>
        </AuthGate>
    );
}
