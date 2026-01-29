import styles from "../page.module.css";
import Sidebar from "@/components/Sidebar/Sidebar";
import AuthGate from "@/components/Auth/AuthGate";

export default function AppLayout({ children }: { children: React.ReactNode }) {
    return (
        <AuthGate>
            <div className={styles.shell}>
                <aside className={styles.nav}>
                    <Sidebar />
                </aside>

                <div className={styles.content}>{children}</div>
            </div>
        </AuthGate>
    );
}