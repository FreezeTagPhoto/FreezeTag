"use client";
import Gallery from "@/components/Gallery/Gallery";
import FileUploadButton from "@/components/UI/FileUploadButton/FileUploadButton";
import { useState } from "react";

export default function Home() {
  const [ids, setIds] = useState<number[]>([]);
  const ids_retrieved_callback = (ids: number[]) => {
    setIds(ids);
  };
  return (
    <div>
      <FileUploadButton ids_retrieved_callback={ids_retrieved_callback} />
      <Gallery image_ids={ids} />
    </div>
  );
}
