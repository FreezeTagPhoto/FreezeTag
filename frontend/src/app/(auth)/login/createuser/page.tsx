import LoginRedirect from "@/components/Auth/LoginRedirect";
import LoginView, { Logo } from "@/components/Login/LoginView";
import styles from "../page.module.css"; // reusing login styles

export default function CreateUserPage() {
    return (
        <main className={styles.screen}>
            <LoginRedirect />
            <div className={styles.logoRow}>
                <Logo />
            </div>
            <div className={styles.cardRow}>
                <LoginView mode="create" />
            </div>
        </main>
    );
}
