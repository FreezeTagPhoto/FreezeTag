"use client";

import { useState, useRef, KeyboardEvent as ReactKeyboardEvent } from "react";
import styles from "./MainGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";
import PreviewWindow from "../PreviewWindow/PreviewWindow";

export type GalleryProps = {
    image_ids: number[];
    onSearchTag?: (tag: string) => void;
};

export default function MainGallery({ image_ids, onSearchTag }: GalleryProps) {
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

    const handleGridKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
        if (selectedId !== null) return;

        const count = image_ids.length;
        if (count === 0) return;

        switch (event.key) {
            case "ArrowRight": {
                event.preventDefault();
                const next =
                    focusedIndex === null
                        ? 0
                        : Math.min(focusedIndex + 1, count - 1);
                setFocusedIndex(next);
                itemRefs.current[next]?.focus();
                break;
            }
            case "ArrowLeft": {
                event.preventDefault();
                const prev =
                    focusedIndex === null
                        ? count - 1
                        : Math.max(focusedIndex - 1, 0);
                setFocusedIndex(prev);
                itemRefs.current[prev]?.focus();
                break;
            }
            case "Escape": {
                event.preventDefault();

                if (focusedIndex !== null) {
                    itemRefs.current[focusedIndex]?.blur();
                    setFocusedIndex(null);
                } else {
                    (event.currentTarget as HTMLElement).blur();
                }
                break;
            }
        }
    };

    return (
        <>
            {/* thumbnails */}
            <div
                className={styles.grid}
                tabIndex={0}
                onKeyDown={handleGridKeyDown}
                aria-label="Photo gallery"
            >
                {image_ids.map((id, index) => (
                    <GalleryImage
                        key={id}
                        id={id}
                        onClick={() => {
                            setSelectedId(id);
                        }}
                        onFocus={() => {
                            setFocusedIndex(index);
                        }}
                        buttonRef={(el) => {
                            itemRefs.current[index] = el;
                        }}
                        selected={false}
                    />
                ))}
            </div>

            {/* fullscreen preview */}
            {selectedId !== null && (
                <PreviewWindow
                    imageIds={image_ids}
                    selectedId={selectedId}
                    onClose={() => setSelectedId(null)}
                    onNavigate={(nextId, nextIndex) => {
                        setSelectedId(nextId);
                        setFocusedIndex(nextIndex);
                    }}
                    onSearchTag={(t) => {
                        onSearchTag?.(t);
                        setSelectedId(null);
                    }}
                />
            )}
        </>
    );
}
