"use client";
import {
    useEffect,
    useState,
    SubmitEvent,
    ChangeEvent,
    KeyboardEvent,
} from "react";
import {
    GetPluginConfig,
    SetPluginConfig,
    PluginConfigField,
} from "@/api/plugins/pluginconfig";
import styles from "./Configuration.module.css";
import { Plugin } from "@/api/plugins/pluginshelpers";
import { Save, Eraser, RotateCcw } from "lucide-react";

export type ConfigProps = {
    onClose: () => void;
    plugin: Plugin;
};

export default function Config({ onClose, plugin }: ConfigProps) {
    const [fields, setFields] = useState<Record<string, PluginConfigField>>({});
    const [formData, setFormData] = useState<Record<string, string>>({});

    const fetchFields = async () => {
        const result = await GetPluginConfig(plugin.name);
        if (result.ok) {
            setFields(result.value);
        } else {
            console.error(`Plugin Config Error! ${result.error.message}`);
        }
    };

    const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
        const { name, value } = event.target;
        setFormData((prevState) => {
            if (fields[name]?.value === value) {
                const { [name]: _, ...newState } = prevState;
                return newState;
            } else {
                return { ...prevState, [name]: value };
            }
        });
    };

    const noEnterSubmit = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Enter") {
            event.preventDefault();
        }
    };

    const resetField = (field: string) => {
        setFormData((prevState) => {
            const { [field]: _, ...newState } = prevState;
            return newState;
        });
    };

    const resetDefaultField = (field: string) => {
        setFormData((prevState) => {
            const newState = { ...prevState, [field]: fields[field].default };
            if (fields[field]?.value === newState[field]) {
                const { [field]: _, ...newerState } = newState;
                return newerState;
            }
            return newState;
        });
    };

    const handleSubmit = async (event: SubmitEvent<HTMLFormElement>) => {
        event.preventDefault();
        const result = await SetPluginConfig(plugin.name, formData);
        if (!result.ok) {
            console.error(`Error saving form: ${result.error}`);
        }
        setFormData({});
        await fetchFields();
    };

    useEffect(() => {
        fetchFields();
    }, []);

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>
                        Configuration for {plugin.friendly_name}
                    </h1>
                </header>
                <form
                    onSubmit={handleSubmit}
                    className={styles.config_container}
                >
                    {Object.entries(fields).map(([name, field]) => (
                        <div key={name} className={styles.plugin_row}>
                            <label
                                className={styles.plugin_item}
                                htmlFor={"configForm$" + name}
                            >
                                {field.name ?? name}
                                {field.description && (
                                    <p>{field.description}</p>
                                )}
                            </label>
                            <input
                                className={`${styles.plugin_item} ${fields[name].default || styles.plugin_onlyone_button}`}
                                type="text"
                                name={name}
                                id={"configForm$" + name}
                                value={
                                    formData[name] ?? fields[name].value ?? ""
                                }
                                placeholder={
                                    field.protected &&
                                    formData[name] === undefined
                                        ? "Read-Protected"
                                        : ""
                                }
                                onChange={handleChange}
                                onKeyDown={noEnterSubmit}
                                autoComplete="off"
                            ></input>
                            <button
                                className={`${styles.plugin_item} ${styles.plugin_item_button}`}
                                disabled={formData[name] === undefined}
                                onClick={() => resetField(name)}
                            >
                                <Eraser className={styles.icon} />
                                <p className={styles.plugin_item_label}>Undo</p>
                            </button>
                            {fields[name].default && (
                                <button
                                    className={`${styles.plugin_item} ${styles.plugin_item_button}`}
                                    disabled={
                                        (formData[name] ??
                                            fields[name].value) ===
                                        fields[name].default
                                    }
                                    onClick={() => resetDefaultField(name)}
                                >
                                    <RotateCcw className={styles.icon} />
                                    <p className={styles.plugin_item_label}>
                                        Set Default
                                    </p>
                                </button>
                            )}
                        </div>
                    ))}
                    {Object.keys(fields).length != 0 ? (
                        <button
                            className={`${styles.plugin_item} ${styles.plugin_item_button} ${styles.submit_button}`}
                            type="submit"
                            disabled={Object.keys(formData).length == 0}
                        >
                            <Save className={styles.icon} />
                            <p className={styles.plugin_item_label}>
                                Save Changes
                            </p>
                        </button>
                    ) : (
                        <p>Nothing to see here</p>
                    )}
                </form>
            </div>
        </div>
    );
}
