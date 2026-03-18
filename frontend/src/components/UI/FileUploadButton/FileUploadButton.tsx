import ImageUploader, { UploadResult } from "@/api/upload/imageuploader";
import styles from "./FileUploadButton.module.css";
import { useRef, useState } from "react";
import { useDropzone } from "react-dropzone";

export type FileUploadProps = {
    job_id_callback: (id: string) => void;
    disabled?: boolean;
};

const FileChangeHandler = (event: React.ChangeEvent<HTMLInputElement>) => {
    // Make sure that we have at least one file so it doesn't completely snafu
    if (event.target.files && event.target.files[0]) {
        if (event.target.form) {
            event.target.form.requestSubmit();
        } else {
            console.error("Form was null for some reason!");
        }
    }
};

export default function FileUploadButton(props: FileUploadProps) {
    const hiddenInputRef = useRef<HTMLInputElement>(null);

    const [uploading, setUploading] = useState<boolean>(false);
    const handleSubmit = (
        event: FormData,
        job_id_callback: (id: string) => void,
    ) => {
        setUploading(true);
        ImageUploader(event).then((result) =>
            handleResult(job_id_callback, result),
        );
    };

    const handleResult = (
        job_id_callback: (id: string) => void,
        result: UploadResult,
    ) => {
        setUploading(false);
        if (result.ok) {
            job_id_callback(result.value);
        } else {
            console.error(
                "Error uploading images (is the backend running?):",
                result.error,
            );
            // TODO: show error to user
        }
    };

    const { getRootProps, getInputProps } = useDropzone({
        onDrop: (incomingFiles) => {
            if (hiddenInputRef.current) {
                // Note the specific way we need to munge the file into the hidden input
                // https://stackoverflow.com/a/68182158/1068446
                const dataTransfer = new DataTransfer();
                incomingFiles.forEach((v) => {
                    dataTransfer.items.add(v);
                });

                hiddenInputRef.current.files = dataTransfer.files;
                hiddenInputRef.current.dispatchEvent(
                    new Event("change", { bubbles: true }),
                );
            }
        },
        accept: { "image/*": [] },
    });

    return (
        <form action={(e) => handleSubmit(e, props.job_id_callback)}>
            <div
                {...getRootProps({
                    className: `${styles.label} ${
                        props.disabled || uploading ? styles.label_disabled : ""
                    }`,
                })}
            >
                {uploading ? "Images Uploading..." : "Upload Images"}
                <input
                    type="file"
                    onChange={FileChangeHandler}
                    multiple
                    required
                    name="file"
                    className={styles.button}
                    id="file-upload"
                    ref={hiddenInputRef}
                    disabled={props.disabled || uploading}
                />
                <input
                    {...getInputProps()}
                    className={styles.button}
                    disabled={props.disabled || uploading}
                />
            </div>
        </form>
    );
}
