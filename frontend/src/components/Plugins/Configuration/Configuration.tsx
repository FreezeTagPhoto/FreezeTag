"use client";
import { useContext, useEffect, useState, SubmitEvent, ChangeEvent, KeyboardEvent } from "react";
import { GetPluginConfig, SetPluginConfig, PluginGetConfigResponse } from "@/api/plugins/pluginconfig";
import styles from "./Configuration.module.css";
import { Plugin } from "@/api/plugins/pluginshelpers";

export type ConfigProps = {
    onClose: () => void
    plugin: Plugin
}

type ConfigField = {
    name: string,
    value?: string,
    protected: boolean
}

function filterConfigFields(response: PluginGetConfigResponse): ConfigField[] {
    let fields: ConfigField[] = [];
    if (response.has("protected_fields")) {
        response.get("protected_fields").forEach((field: string) => {
            fields.push({
                name: field,
                protected: true
            });
        });
        response.delete("protected_fields")
    }
    response.forEach((value, field) => {
        fields.push({
            name: field,
            protected: false,
            value: value
        });
    });
    return fields;
}

export default function Config({onClose, plugin}: ConfigProps) {
    const [fields, setFields] = useState<ConfigField[]>([]);
    const [formData, setFormData] = useState({});
    
    const fetchFields = async () => {
        const result = await GetPluginConfig(plugin.name);
        if (result.ok) {
            setFields(filterConfigFields(result.value))
        } else {
            console.error(`Plugin Config Error! ${result.error.message}`);
        }
    }

    const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
        const { name, value } = event.target;
        setFormData((prevState) => ({
            ...prevState,
            [name]: value,
        }));
    }

    const noEnterSubmit = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Enter') {
            event.preventDefault();
        }
    }

    const handleSubmit = async (event: SubmitEvent<HTMLFormElement>) => {
        event.preventDefault()
        const result = await SetPluginConfig(plugin.name, formData)
        if (!result.ok) {
            console.error(`Error saving form: ${result.error}`)
        }
    }

    useEffect(() => {
        fetchFields();
    }, [])

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Configuration for {plugin.friendly_name}</h1>
                </header>
                <div className={styles.config_container}>
                    <form onSubmit={handleSubmit}>
                    {fields.map(field => (
                        <div key={field.name}>
                            <label htmlFor={"configForm$" + field.name}>{field.name}:</label>
                            <input type="text" name={field.name} id={"configForm$" + field.name} defaultValue={field.value} placeholder={field.protected ? "Read-Protected" : ""} onChange={handleChange} onKeyDown={noEnterSubmit} autoComplete="off"></input>
                        </div>
                    ))}
                    {fields.length != 0 ?
                        <button type="submit" disabled={Object.keys(formData).length == 0}>Save Changes</button>
                        :
                        <p>Nothing to see here</p>
                    }
                    </form>
                </div>
            </div>
        </div>
    )
}