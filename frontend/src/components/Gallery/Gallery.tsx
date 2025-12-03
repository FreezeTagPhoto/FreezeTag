"use client";

import {
    useEffect,
    useState,
    useRef,
    MouseEvent,
    KeyboardEvent as ReactKeyboardEvent,
    useCallback,
} from "react";
import styles from "./Gallery.module.css";
import GalleryImage from "./GalleryImage/GalleryImage";

export type GalleryProps = {
    image_ids: number[];
    selectable_images?: boolean;
    onChange?: (ids: Set<number>) => void;
};

export default function Gallery({
    image_ids,
    selectable_images,
    onChange,
}: GalleryProps) {
    // Full Screen Preview Handling
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

    const gridRef = useRef<HTMLDivElement | null>(null);
    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);
    const viewerImageAreaRef = useRef<HTMLDivElement | null>(null);

    // zoom: 1 = fit, 2 = zoomed
    const [zoom, setZoom] = useState<number>(1);

    const moveSelection = useCallback(
        (direction: "next" | "prev") => {
            if (selectedId === null) return;

            const currentIndex = image_ids.indexOf(selectedId);
            if (currentIndex === -1) return;

            const delta = direction === "next" ? 1 : -1;
            const nextIndex = currentIndex + delta;

            if (nextIndex < 0 || nextIndex >= image_ids.length) return;

            const nextId = image_ids[nextIndex];
            setSelectedId(nextId);
            setFocusedIndex(nextIndex);
        },
        [selectedId, image_ids],
    );

    // reset zoom + scroll when switching / closing images
    useEffect(() => {
        setZoom(1);
        const area = viewerImageAreaRef.current;
        if (area) {
            area.scrollLeft = 0;
            area.scrollTop = 0;
        }
    }, [selectedId]);

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

            if (event.key === "ArrowRight") {
                moveSelection("next");
            } else if (event.key === "ArrowLeft") {
                moveSelection("prev");
            }
        };

        window.addEventListener("keydown", handleKeyDown, true);
        return () => window.removeEventListener("keydown", handleKeyDown, true);
    }, [selectedId, moveSelection]);

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

    const zoomOut = () => {
        setZoom(1);
        const area = viewerImageAreaRef.current;
        if (!area) return;
        requestAnimationFrame(() => {
            area.scrollLeft = 0;
            area.scrollTop = 0;
        });
    };

    const zoomInCentered = () => {
        const area = viewerImageAreaRef.current;
        setZoom(2);
        if (!area) return;
        requestAnimationFrame(() => {
            const maxLeft = area.scrollWidth - area.clientWidth;
            const maxTop = area.scrollHeight - area.clientHeight;
            area.scrollLeft = Math.max(0, maxLeft / 2);
            area.scrollTop = Math.max(0, maxTop / 2);
        });
    };

    const handleZoomButtonClick = (event: MouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        if (zoom === 1) {
            zoomInCentered();
        } else {
            zoomOut();
        }
    };

    // click-to-zoom at cursor position
    const handleImageAreaClick = (event: MouseEvent<HTMLDivElement>) => {
        const target = event.target as HTMLElement;
        // don't toggle zoom when clicking buttons over the image
        if (target.closest("button")) return;

        const area = viewerImageAreaRef.current;
        if (!area) return;

        if (zoom === 1) {
            const rect = area.getBoundingClientRect();

            // position of click within the scrollable content (before zoom)
            const offsetX = event.clientX - rect.left + area.scrollLeft;
            const offsetY = event.clientY - rect.top + area.scrollTop;

            const newZoom = 2;
            const scaleFactor = newZoom / zoom; // 2 / 1

            setZoom(newZoom);

            // after layout updates for the new zoom, adjust scroll so the
            // clicked point is roughly centered
            requestAnimationFrame(() => {
                const desiredLeft =
                    offsetX * scaleFactor - area.clientWidth / 2;
                const desiredTop =
                    offsetY * scaleFactor - area.clientHeight / 2;

                const maxLeft = area.scrollWidth - area.clientWidth;
                const maxTop = area.scrollHeight - area.clientHeight;

                area.scrollLeft = Math.max(0, Math.min(desiredLeft, maxLeft));
                area.scrollTop = Math.max(0, Math.min(desiredTop, maxTop));
            });
        } else {
            // currently zoomed -> zoom back out
            zoomOut();
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
                        onClick={() => {
                            if (selectable_images) {
                                handleImageSelectionChange(id);
                            } else {
                                setSelectedId(id);
                            }
                        }}
                        onFocus={() => {
                            // Avoids focusing in select mode
                            if (!selectable_images) setFocusedIndex(index);
                        }}
                        buttonRef={(el) => {
                            itemRefs.current[index] = el;
                        }}
                        selected={selectedIds.has(id)}
                    />
                ))}
            </div>

            {/* fullscreen previews */}
            {selectedId !== null && (
                <div
                    className={styles.viewerBackdrop}
                    onClick={handleBackdropClick}
                >
                    <div className={styles.viewer} onClick={stopPropagation}>
                        <div
                            className={styles.viewerImageArea}
                            ref={viewerImageAreaRef}
                            onClick={handleImageAreaClick}
                            style={{
                                cursor: zoom === 1 ? "zoom-in" : "zoom-out",
                            }}
                        >
                            {/* chevron buttons */}
                            <button
                                type="button"
                                className={`${styles.navButton} ${styles.navButtonLeft}`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    moveSelection("prev");
                                }}
                                aria-label="Previous image"
                            >
                                ‹
                            </button>
                            <button
                                type="button"
                                className={`${styles.navButton} ${styles.navButtonRight}`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    moveSelection("next");
                                }}
                                aria-label="Next image"
                            >
                                ›
                            </button>
                            {/* close button */}
                            <button
                                type="button"
                                className={styles.closeButton}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedId(null);
                                }}
                                aria-label="Close"
                            >
                                ×
                            </button>

                            {/* zoom toggle button (1x / 2x) */}
                            <button
                                type="button"
                                className={styles.zoomButton}
                                onClick={handleZoomButtonClick}
                                aria-label={
                                    zoom === 1 ? "Zoom to 2x" : "Zoom to 1x"
                                }
                            >
                                {zoom === 1 ? "1x" : "2x"}
                            </button>

                            <img
                                src={`http://localhost:3824/thumbnails/${selectedId}?size=2`}
                                alt={`Preview of image ${selectedId}`}
                                className={styles.viewerImage}
                                style={
                                    zoom === 1
                                        ? {
                                              width: "100%",
                                              height: "auto",
                                              maxHeight: "100%",
                                          }
                                        : {
                                              width: `${zoom * 100}%`,
                                              height: "auto",
                                          }
                                }
                            />
                        </div>

                        <aside className={styles.viewerSidebar}>
                            <h2 className={styles.sidebarTitle}>
                                Image details
                            </h2>
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
