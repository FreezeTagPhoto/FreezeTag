import styles from "./Gallery.module.css";
import GalleryImage from "./GalleryImage/GalleryImage";

export type GalleryProps = {
  image_ids: number[];
};

export default function Gallery(props: GalleryProps) {
  return (
    <div className={styles.grid}>
      {props.image_ids.map((id) => (
        <GalleryImage key={id} id={id} />
      ))}
    </div>
  );
}
