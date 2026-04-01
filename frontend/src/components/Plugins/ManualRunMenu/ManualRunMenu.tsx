import { useEffect, useState } from "react";
import styles from "./ManualRunMenu.module.css";
import PluginsLister from "@/api/plugins/pluginslister";
import { Play } from "lucide-react";

export type ManualRunMenuProps = {
    onClose: () => void;
    onPluginChosen: (
        plugin_name: string,
        hook_name: string,
        hook_signature: string,
        hook_type: string,
        form_receive_hook_name?: string,
    ) => void;
    multipleImages: boolean;
};

type PluginHookPair = {
    plugin_name: string;
    plugin_friendly_name: string;
    hook_name: string;
    hook_friendly_name: string;
    hook_signature: string;
    hook_type: "manual_trigger" | "generate_form";
    form_receive_hook_name?: string;
};

export default function ManualRunMenu({
    onClose,
    onPluginChosen,
    multipleImages,
}: ManualRunMenuProps) {
    const [plugins, setPlugins] = useState<PluginHookPair[] | undefined>(
        undefined,
    );

    useEffect(() => {
        (async () => {
            const result = await PluginsLister();
            if (!result.ok) {
                console.error(
                    `plugin list error: ${JSON.stringify(result.error)}`,
                );
                return;
            }
            const plugins = result.value;

            const pairs: PluginHookPair[] = [];

            for (const plugin of plugins) {
                Object.entries(plugin.hooks).forEach(([name, hook]) => {
                    if (hook.signature === "form_data") {
                        return;
                    }
                    if (multipleImages && hook.signature === "single_image") {
                        return;
                    }
                    if (hook.type === "manual_trigger") {
                        pairs.push({
                            plugin_friendly_name: plugin.friendly_name,
                            plugin_name: plugin.name,
                            hook_name: name,
                            hook_friendly_name: hook.friendly_name,
                            hook_signature: hook.signature,
                            hook_type: hook.type,
                        });
                    } else if (hook.type === "generate_form") {
                        const r = Object.entries(plugin.hooks).find(
                            ([_name, hook]) => hook.signature === "form_data",
                        );
                        if (!r) {
                            console.error(
                                "plugins with generate_form hooks should have a form_data hook!",
                            );
                            return;
                        }
                        pairs.push({
                            plugin_friendly_name: plugin.friendly_name,
                            plugin_name: plugin.name,
                            hook_name: name,
                            hook_friendly_name: hook.friendly_name,
                            hook_signature: hook.signature,
                            hook_type: hook.type,
                            form_receive_hook_name: r[0],
                        });
                    }
                });
            }
            setPlugins(
                pairs.toSorted((a, b) =>
                    a.plugin_friendly_name.localeCompare(
                        b.plugin_friendly_name,
                    ),
                ),
            );
        })();
    }, [multipleImages]);

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Apply Plugins to Group</h1>
                </header>
                <div className={styles.perms_container}>
                    {plugins !== undefined &&
                        plugins.map((pair) => (
                            <div
                                key={`${pair.plugin_name}-${pair.hook_name}`}
                                className={`${styles.perms_row}`}
                            >
                                <button
                                    type="button"
                                    className={`${styles.perms_item} ${styles.plugin_item_button}`}
                                    onClick={() =>
                                        onPluginChosen(
                                            pair.plugin_name,
                                            pair.hook_name,
                                            pair.hook_signature,
                                            pair.hook_type,
                                            pair.form_receive_hook_name,
                                        )
                                    }
                                    title="Run Plugin"
                                >
                                    <Play className={`${styles.icon}`} />
                                </button>
                                <div
                                    className={`${styles.perms_item} ${styles.text}`}
                                    title={pair.plugin_friendly_name}
                                >
                                    <p className={styles.text_preview}>
                                        {pair.plugin_friendly_name}
                                    </p>
                                </div>
                                <div
                                    className={`${styles.perms_item} ${styles.text}`}
                                    title={pair.hook_friendly_name}
                                >
                                    <p className={styles.text_preview}>
                                        {pair.hook_friendly_name}
                                    </p>
                                </div>
                            </div>
                        ))}
                    {plugins !== undefined && plugins.length === 0 && (
                        <p>
                            No plugins available for the input type of{" "}
                            {multipleImages
                                ? "multiple images"
                                : "single image"}
                        </p>
                    )}
                </div>
            </div>
        </div>
    );
}
