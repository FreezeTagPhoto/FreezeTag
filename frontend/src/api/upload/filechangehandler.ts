"use client";
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

export default FileChangeHandler;
