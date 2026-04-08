"use client";
import { useContext, useEffect, useState } from "react";
import styles from "./page.module.css";
import { UserContext } from "@/components/Auth/AuthGate";
import { Plugin } from "@/api/plugins/pluginshelpers";
import PluginsLister from "@/api/plugins/pluginslister";
import { UserHasPerm } from "@/api/permissions/permshelpers";
import PluginsAbler from "@/api/plugins/pluginsabler";
import { FishingHook, History, Power, Settings, Upload } from "lucide-react";
import Hooks from "@/components/Plugins/Hooks/Hooks";
import Config from "@/components/Plugins/Configuration/Configuration";
import UploadPlugin from "@/components/Plugins/UploadPlugin/UploadPlugin";

export default function Home() {
    const [plugins, setPlugins] = useState<Plugin[]>([]);

    const [viewingHooks, setViewingHooks] = useState<Plugin | undefined>();
    const [viewingConfig, setViewingConfig] = useState<Plugin | undefined>();

    const currentUser = useContext(UserContext);
    const userCanChangePlugins = UserHasPerm(currentUser, "write:plugins");

    const [uploadingPlugin, setUploadingPlugin] = useState<boolean>(false);

    const fetchPlugins = async () => {
        const result = await PluginsLister();
        if (result.ok) {
            setPlugins(
                result.value.sort((a, b) => a.name.localeCompare(b.name)),
            );
        } else {
            console.error(`Plugin Lister Error! ${result.error.message}`);
        }
    };

    useEffect(() => {
        fetchPlugins();
    }, [uploadingPlugin]);

    const onPluginAble = async (
        current_state: boolean,
        plugin_name: string,
    ) => {
        const result = await PluginsAbler(plugin_name, !current_state);
        if (result.ok) {
            fetchPlugins();
        } else {
            console.error(`Plugin Abler Error! ${result.error}`);
        }
    };

    return (
        <main className={styles.main}>
            <h1 className={styles.h1}>Plugin Management</h1>
            <div className={styles.plugin_container}>
                {plugins.map((plugin) => (
                    <div key={plugin.name} className={styles.plugin_row}>
                        <button
                            type="button"
                            className={`${styles.plugin_item} ${styles.plugin_item_button}`}
                            disabled={!userCanChangePlugins}
                            onClick={() =>
                                onPluginAble(plugin.enabled, plugin.name)
                            }
                            title={
                                userCanChangePlugins
                                    ? plugin.enabled
                                        ? "Click to disable"
                                        : "Click to enable"
                                    : plugin.enabled
                                      ? "Enabled"
                                      : "Disabled"
                            }
                        >
                            <Power
                                className={`${styles.icon} ${plugin.enabled ? styles.power_icon_on : styles.power_icon_off}`}
                            />
                        </button>
                        <div
                            className={`${styles.plugin_item} ${styles.plugin_name}`}
                            title={plugin.friendly_name}
                        >
                            <p className={styles.plugin_name_preview}>
                                {plugin.friendly_name}
                            </p>
                        </div>
                        <p
                            className={`${styles.plugin_item} ${styles.version}`}
                            title={plugin.version}
                        >
                            <History className={styles.icon} />
                            {plugin.version}
                        </p>
                        <button
                            type="button"
                            className={`${styles.plugin_item} ${styles.plugin_item_button}`}
                            onClick={() => setViewingHooks(plugin)}
                        >
                            <FishingHook className={styles.icon} />
                            <p className={styles.plugin_item_label}>Hooks</p>
                        </button>
                        {plugin.configurable && (
                            <button
                                type="button"
                                className={`${styles.plugin_item} ${styles.plugin_item_button}`}
                                onClick={() => setViewingConfig(plugin)}
                            >
                                <Settings className={styles.icon} />
                                <p className={styles.plugin_item_label}>
                                    Configure
                                </p>
                            </button>
                        )}
                    </div>
                ))}
            </div>
            {viewingHooks && (
                <Hooks
                    onClose={() => setViewingHooks(undefined)}
                    plugin={viewingHooks}
                />
            )}
            {viewingConfig && (
                <Config
                    onClose={() => setViewingConfig(undefined)}
                    plugin={viewingConfig}
                />
            )}
            <button
                type="button"
                className={styles.upload_plugin}
                disabled={!userCanChangePlugins}
                onClick={() => setUploadingPlugin(true)}
            >
                <Upload aria-hidden={true} />
                Upload Plugin
            </button>
            {uploadingPlugin && (
                <UploadPlugin onClose={() => setUploadingPlugin(false)} />
            )}
        </main>
    );
}
