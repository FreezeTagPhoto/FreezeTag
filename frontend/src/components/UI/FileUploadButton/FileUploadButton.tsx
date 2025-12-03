import ImageUploader from "@/api/upload/imageuploader";
import FileChangeHandler from "@/api/upload/filechangehandler";
import styles from "./FileUploadButton.module.css";
import { useRef } from "react";
import { useDropzone } from "react-dropzone";

export type FileUploadProps = {
    ids_retrieved_callback: (ids: number[]) => void;
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
        <form action={(e) => handleSubmit(e, props.ids_retrieved_callback)}>
            <label {...getRootProps({ className: styles.label })}>
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
                />
                <input {...getInputProps()} />
            </label>
        </form>
    );
}

const handleSubmit = async (
    event: FormData,
    ids_retrieved_callback: (ids: number[]) => void,
) => {
    try {
        const result = await ImageUploader(event);

        const ids = [];
        if (result.ok) {
            for (const [key, value] of result.value) {
                if (value.ok) {
                    ids.push(value.value);
                } else {
                    console.error(
                        `Error uploading ${key} because of ${value.error}`,
                    );
                }
            }
        } else {
            console.error(
                "Error uploading images (is the backend running?):",
                result.error,
            );
            // TODO: show error to user
        }

        ids_retrieved_callback(ids);
    } catch (error) {
        console.error(
            "Error uploading images (is the backend running?):",
            error,
        );
        // TODO: show error to user
    }
};
