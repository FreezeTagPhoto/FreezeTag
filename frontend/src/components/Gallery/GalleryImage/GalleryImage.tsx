import styles from "./GalleryImage.module.css";

export type GalleryImageProps = {
  id: number;
  onClick?: () => void;
};

export default function GalleryImage({ id, onClick }: GalleryImageProps) {
  return (
    <button
      type="button"
      className={styles.image_container}
      onClick={onClick}
    >
      <img
        src={`http://localhost:3824/thumbnails/${id}?size=1`}
        loading="lazy"
        alt={`A thumbnail of image ${id}`}
        decoding="async"
        className={styles.image}
      />
    </button>
  );
}
