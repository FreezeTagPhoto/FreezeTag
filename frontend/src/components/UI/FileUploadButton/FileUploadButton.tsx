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
        File Upload!{" "}
      </label>
      <input
        type="file"
        onChange={FileChangeHandler}
        multiple
        required
        name="file"
        className={styles.button}
        id="file-upload"
      />
    </form>
  );
}

const handleSubmit = async (
  event: FormData,
  ids_retrieved_callback: (ids: number[]) => void
) => {
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
};
