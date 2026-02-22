"use client";
import { useContext, useEffect, useState } from "react";
import styles from "./page.module.css";
import { UserContext } from "@/components/Auth/AuthGate";
import { Plugin } from "@/api/plugins/pluginshelpers";
import PluginsLister from "@/api/plugins/pluginslister";

export default function Home() {
    const [plugins, setPlugins] = useState<Plugin[]>([]);

    const currentUser = useContext(UserContext);

    useEffect(() => {
        (async () => {
            const result = await PluginsLister();
            if (result.ok) {
                setPlugins(result.value);
            } else {
                console.error(`Plugin Lister Error! ${result.error.message}`);
            }
        })();
    }, []);

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Plugin Management</h1>
            <div className={styles.user_container}>
                {plugins.map((plugin) => (
                    <div key={plugin.name} className={styles.account_row}>
                        <div
                            className={`${styles.account_item} ${styles.username}`}
                            title={plugin.name}
                        >
                            <p className={styles.username_preview}>
                                {plugin.name}
                            </p>
                        </div>
                        <p
                            className={`${styles.account_item} ${styles.date}`}
                            title={plugin.version}
                        >
                            {plugin.version}
                        </p>
                    </div>
                ))}
            </div>
        </main>
    );
}
