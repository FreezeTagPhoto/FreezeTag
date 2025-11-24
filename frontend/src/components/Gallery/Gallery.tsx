"use client";

import { useEffect, useState, MouseEvent } from "react";
import styles from "./Gallery.module.css";
import GalleryImage from "./GalleryImage/GalleryImage";

export type GalleryProps = {
  image_ids: number[];
};

export default function Gallery({ image_ids }: GalleryProps) {
  const [selectedId, setSelectedId] = useState<number | null>(null);

  // esc key works
  useEffect(() => {
    if (selectedId === null) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setSelectedId(null);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [selectedId]);

  const handleBackdropClick = () => {
    setSelectedId(null);
  };

  const stopPropagation = (e: MouseEvent<HTMLDivElement>) => {
    e.stopPropagation();
  };

  return (
    <>
      {/* thumbnails */}
      <div className={styles.grid}>
        {image_ids.map((id) => (
          <GalleryImage key={id} id={id} onClick={() => setSelectedId(id)} />
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
