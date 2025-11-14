import ImageUploader from "@/api/upload/imageuploader";
import FileChangeHandler from "@/api/upload/filechangehandler";
import styles from "./FileUploadButton.module.css";

export default function FileUploadButton() {
  return (
    <form action={handleSubmit}>
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

const handleSubmit = async (event: FormData) => {
  "use server";
  const result = await ImageUploader(event);
  console.log(result);
};
