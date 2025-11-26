"use client";

import {
  useEffect,
  useState,
  useRef,
  MouseEvent,
  KeyboardEvent as ReactKeyboardEvent,
} from "react";
import styles from "./Gallery.module.css";
import GalleryImage from "./GalleryImage/GalleryImage";

export type GalleryProps = {
  image_ids: number[];
};

export default function Gallery({ image_ids }: GalleryProps) {
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

  const gridRef = useRef<HTMLDivElement | null>(null);
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

  useEffect(() => {
    if (selectedId === null) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        setSelectedId(null);
        return;
      }

      if (event.key !== "ArrowRight" && event.key !== "ArrowLeft") {
        return;
      }

      event.preventDefault();
      event.stopPropagation();

      const currentIndex = image_ids.indexOf(selectedId);
      if (currentIndex === -1) return;

      let nextIndex = currentIndex;

      if (event.key === "ArrowRight") {
        if (currentIndex < image_ids.length - 1) {
          nextIndex = currentIndex + 1;
        } else {
          return;
        }
      } else if (event.key === "ArrowLeft") {
        if (currentIndex > 0) {
          nextIndex = currentIndex - 1;
        } else {
          return;
        }
      }

      const nextId = image_ids[nextIndex];
      setSelectedId(nextId);
      setFocusedIndex(nextIndex);
    };

    window.addEventListener("keydown", handleKeyDown, true);
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [selectedId, image_ids]);

  const handleBackdropClick = () => {
    setSelectedId(null);
  };

  const stopPropagation = (e: MouseEvent<HTMLDivElement>) => {
    e.stopPropagation();
  };

  const handleGridKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (selectedId !== null) return;

    const count = image_ids.length;
    if (count === 0) return;

    switch (event.key) {
      case "ArrowRight": {
        event.preventDefault();
        const next =
          focusedIndex === null ? 0 : Math.min(focusedIndex + 1, count - 1);
        setFocusedIndex(next);
        itemRefs.current[next]?.focus();
        break;
      }
      case "ArrowLeft": {
        event.preventDefault();
        const prev =
          focusedIndex === null ? count - 1 : Math.max(focusedIndex - 1, 0);
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
        ref={gridRef}
        tabIndex={0}
        onKeyDown={handleGridKeyDown}
        aria-label="Photo gallery"
      >
        {image_ids.map((id, index) => (
          <GalleryImage
            key={id}
            id={id}
            onClick={() => setSelectedId(id)}
            onFocus={() => setFocusedIndex(index)}
            buttonRef={(el) => {
              itemRefs.current[index] = el;
            }}
          />
        ))}
      </div>

      {/* fullscreen previews */}
      {selectedId !== null && (
        <div className={styles.viewerBackdrop} onClick={handleBackdropClick}>
          <div className={styles.viewer} onClick={stopPropagation}>
            <div className={styles.viewerImageArea}>
              <button
                type="button"
                className={styles.closeButton}
                onClick={() => setSelectedId(null)}
                aria-label="Close"
              >
                ×
              </button>

              <img
                src={`http://localhost:3824/thumbnails/${selectedId}?size=2`}
                alt={`Preview of image ${selectedId}`}
                className={styles.viewerImage}
              />
            </div>

            <aside className={styles.viewerSidebar}>
              <h2 className={styles.sidebarTitle}>Image details</h2>
              <dl className={styles.sidebarList}>
                {/* //TODO: these are placeholders, swap for real metadata later */}
                <div>
                  <dt>Filename</dt>
                  <dd>IMG_{selectedId}.JPG</dd>
                </div>
                <div>
                  <dt>Date taken</dt>
                  <dd>Jun 7, 6:41 PM</dd>
                </div>
                <div>
                  <dt>Location</dt>
                  <dd>Salt Lake City</dd>
                </div>
                <div>
                  <dt>Camera</dt>
                  <dd>Apple iPhone 16 Pro</dd>
                </div>
                <div>
                  <dt>Resolution</dt>
                  <dd>6767 x 4141</dd>
                </div>
              </dl>
            </aside>
          </div>
        </div>
      )}
    </>
  );
}
