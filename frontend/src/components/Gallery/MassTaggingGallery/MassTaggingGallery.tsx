"use client";

import { useRef } from "react";
import styles from "./MassTaggingGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";

export type MassTaggingGalleryProps = {
    image_ids: number[];
    selectedIds: Set<number>;
    onChange: (id: number) => void;
};

export default function MassTaggingGallery({
    image_ids,
    selectedIds,
    onChange,
}: MassTaggingGalleryProps) {
    const gridRef = useRef<HTMLDivElement | null>(null);
    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

    return (
        <>
            {/* thumbnails */}
            <div
                className={styles.grid}
                ref={gridRef}
                aria-label="Photo gallery"
            >
                {image_ids.map((id, index) => (
                    <GalleryImage
                        key={id}
                        id={id}
                        onClick={() => onChange(id)}
                        buttonRef={(el) => {
                            itemRefs.current[index] = el;
                        }}
                        selected={selectedIds.has(id)}
                    />
                ))}
            </div>
        </>
    );
}
