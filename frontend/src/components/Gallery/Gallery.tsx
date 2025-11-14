import styles from "./Gallery.module.css";
import Image from "next/image";

type Item = {
  id: string;
  colSpan: number;
  rowSpan: number;
};

const items: Item[] = [
  { id: "1", colSpan: 3, rowSpan: 2 },
  { id: "2", colSpan: 2, rowSpan: 2 },
  { id: "3", colSpan: 3, rowSpan: 2 },
  { id: "4", colSpan: 3, rowSpan: 2 },
  { id: "5", colSpan: 5, rowSpan: 3 },
  { id: "6", colSpan: 3, rowSpan: 2 },
  { id: "7", colSpan: 3, rowSpan: 2 },
  { id: "8", colSpan: 6, rowSpan: 2 },
  { id: "9", colSpan: 6, rowSpan: 2 },
  { id: "10", colSpan: 6, rowSpan: 2 },
];

export type GalleryProps = {
  image_ids: number[];
};

export default function Gallery(props: GalleryProps) {
  if (props.image_ids) {
    return (
      <div className={styles.grid}>
        {props.image_ids.map((id) => (
          <Image
            key={id}
            src={`http://localhost:3824/thumbnails/${id}?size=1`}
            alt={`A thumbnail of image ${id}`}
            width={512}
            height={512}
            loading="lazy"
          />
        ))}
      </div>
    );
  }
  return (
    <div className={styles.grid}>
      {items.map((it) => (
        <div
          key={it.id}
          className={styles.tile}
          style={{
            gridColumn: `span ${it.colSpan}`,
            gridRow: `span ${it.rowSpan}`,
          }}
        >
          <div className={styles.thumb}>
            <div className={styles.shapeTri} />
            <div className={styles.shapeGear} />
            <div className={styles.shapeSquare} />
          </div>
        </div>
      ))}
    </div>
  );
}
