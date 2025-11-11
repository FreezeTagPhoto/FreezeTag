import ImageUploader from "@/api/upload/imageuploader";
import FileChangeHandler from "@/api/upload/filechangehandler";

export default function FileUploadButton() {
  return (
    <form action={handleSubmit} method="POST">
      <input
        type="file"
        onChange={FileChangeHandler}
        multiple
        required
        accept="image/png"
        name="image"
      />
    </form>
  );
}

const handleSubmit = async (event: FormData) => {
  "use server";
  const result = await ImageUploader(event);
  console.log(result);
};
