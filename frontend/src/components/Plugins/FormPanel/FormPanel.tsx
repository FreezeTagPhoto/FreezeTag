import { useEffect, useState } from "react";
import styles from "./FormPanel.module.css";
import FormDataPluginRunner from "@/api/plugins/formdatapluginrunner";

export type FormPanelProps = {
    formString: string;
    plugin: string;
    hook: string;
    onClose: () => void;
    onFormSubmit: () => void;
};

export default function FormPanel({
    formString,
    plugin,
    hook,
    onClose,
    onFormSubmit,
}: FormPanelProps) {
    const [processedString, setProcessedString] = useState<string>("");

    useEffect(() => {
        const formTagLen = "<form>".length;
        const formCloseTagLen = "</form>".length;
        const lastIndex = formString.length - formCloseTagLen;

        setProcessedString(formString.substring(formTagLen, lastIndex));
    }, [formString]);

    const onSubmit = (e: FormData) => {
        FormDataPluginRunner(plugin, hook, e);
        onFormSubmit();
    };

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
                <header className={styles.header}>
                    <h1 className={styles.h1}>Complete Form</h1>
                </header>
                <form
                    action={onSubmit}
                    dangerouslySetInnerHTML={{ __html: processedString }}
                ></form>
            </div>
        </div>
    );
}
