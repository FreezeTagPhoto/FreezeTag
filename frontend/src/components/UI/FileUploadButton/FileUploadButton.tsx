import ImageUploader from "@/api/upload/imageuploader";
import FileChangeHandler from "@/api/upload/filechangehandler";
import styles from "./FileUploadButton.module.css";

export type FileUploadProps = {
    ids_retrieved_callback: (ids: number[]) => void;
};

export default function FileUploadButton(props: FileUploadProps) {
    return (
        <form action={(e) => handleSubmit(e, props.ids_retrieved_callback)}>
            <label htmlFor="file-upload" className={styles.label}>
                {" "}
                Upload images{" "}
            </label>
            <input
                type="file"
                onChange={FileChangeHandler}
                multiple
                required
                name="file"
                className={styles.button}
                id="file-upload"
                accept="image/*"
            />
        </form>
    );
}

const handleSubmit = async (
    event: FormData,
    ids_retrieved_callback: (ids: number[]) => void,
) => {
    try {
        const result = await ImageUploader(event);
        console.log(result);

        const ids = [];
        if (result.ok) {
            for (const [_key, value] of result.value) {
                if (value.ok) {
                    ids.push(value.value);
                }
            }
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
