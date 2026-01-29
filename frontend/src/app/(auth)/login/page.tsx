import LoginRedirect from "@/components/Auth/LoginRedirect";
import LoginView, { Logo } from "@/components/Login/LoginView";
import styles from "./page.module.css";

export default function LoginPage() {
    return (
        <main className={styles.screen}>
            <LoginRedirect />
            <div className={styles.logoRow}>
                <Logo />
            </div>
            <div className={styles.cardRow}>
                <LoginView mode="login" />
            </div>
        </main>
    );
}