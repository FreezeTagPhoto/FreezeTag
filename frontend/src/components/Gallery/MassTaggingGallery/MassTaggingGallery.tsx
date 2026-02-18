"use client";

import { useState, useRef } from "react";
import styles from "./MassTaggingGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";

export type MassTaggingGalleryProps = {
    image_ids: number[];
    onChange: (ids: Set<number>) => void;
};

export default function MassTaggingGallery({
    image_ids,
    onChange,
}: MassTaggingGalleryProps) {
    const gridRef = useRef<HTMLDivElement | null>(null);
    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

    // Selectable Images Handling
    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

    const handleImageSelectionChange = (id: number) => {
        const newIds = new Set<number>().union(selectedIds);

        if (newIds.has(id)) {
            newIds.delete(id);
        } else {
            newIds.add(id);
        }

        setSelectedIds(newIds);

        if (onChange) {
            onChange(newIds);
        }
    };

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
                        onClick={() => {
                            handleImageSelectionChange(id);
                        }}
                        onFocus={() => {}}
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
