import styles from "./GalleryImage.module.css";

export type GalleryImageProps = {
  id: number;
};

export default function GalleryImage(props: GalleryImageProps) {
  return (
    <div className={styles.image_container}>
      <img
        src={`http://localhost:3824/thumbnails/${props.id}?size=1`}
        loading="lazy"
        alt={`A thumbnail of image ${props.id}`}
        decoding="async"
        className={styles.image}
      />
    </div>
  );
}
