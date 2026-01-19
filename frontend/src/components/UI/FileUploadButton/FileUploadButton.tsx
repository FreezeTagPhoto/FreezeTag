import ImageUploader from "@/api/upload/imageuploader";
import FileChangeHandler from "@/api/upload/filechangehandler";
import styles from "./FileUploadButton.module.css";
import { useRef } from "react";
import { useDropzone } from "react-dropzone";

export type FileUploadProps = {
    job_id_callback: (id: string) => void;
    disabled?: boolean;
};

export default function FileUploadButton(props: FileUploadProps) {
    const hiddenInputRef = useRef<HTMLInputElement>(null);

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
                    className: props.disabled
                        ? styles.label_disabled
                        : styles.label,
                })}
            >
                {" "}
                Upload images{" "}
                <input
                    type="file"
                    onChange={FileChangeHandler}
                    multiple
                    required
                    name="file"
                    className={styles.button}
                    id="file-upload"
                    ref={hiddenInputRef}
                    disabled={props.disabled}
                />
                <input
                    {...getInputProps()}
                    className={styles.button}
                    disabled={props.disabled}
                />
            </div>
        </form>
    );
}

const handleSubmit = async (
    event: FormData,
    job_id_callback: (id: string) => void,
) => {
    try {
        const result = await ImageUploader(event);

        if (result.ok) {
            job_id_callback(result.value);
        } else {
            console.error(
                "Error uploading images (is the backend running?):",
                result.error,
            );
            // TODO: show error to user
        }
    } catch (error) {
        console.error(
            "Error uploading images (is the backend running?):",
            error,
        );
        // TODO: show error to user
    }
};
