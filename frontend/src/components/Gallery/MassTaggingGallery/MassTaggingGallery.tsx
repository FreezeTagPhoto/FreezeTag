"use client";

import {
    useEffect,
    useState,
    useRef,
    MouseEvent,
    KeyboardEvent as ReactKeyboardEvent,
    useCallback,
} from "react";
import styles from "./MassTaggingGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";

export type GalleryProps = {
    image_ids: number[];
    selectable_images?: boolean;
    onChange?: (ids: Set<number>) => void;
};

// point (fx, fy) on image expressed as fraction of width/height (after zoom)
type PendingPan = null | { fx: number; fy: number };

export default function MainGallery({
    image_ids,
    selectable_images,
    onChange,
}: GalleryProps) {
    // Full Screen Preview Handling
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

    const gridRef = useRef<HTMLDivElement | null>(null);
    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

    // zoom: 1 = fit, 2 = zoomed
    const [zoom, setZoom] = useState<number>(1);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const [hoveringImage, setHoveringImage] = useState(false);

    // remember where to pan after zoom has been applied
    const [pendingPan, setPendingPan] = useState<PendingPan>(null);

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

    // reset zoom + pan when opening / changing image
    useEffect(() => {
        setZoom(1);
        setPendingPan(null);
        const scroller = scrollRef.current;
        if (scroller) {
            scroller.scrollLeft = 0;
            scroller.scrollTop = 0;
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
        setPendingPan(null);
        const scroller = scrollRef.current;
        if (!scroller) return;

        scroller.scrollLeft = 0;
        scroller.scrollTop = 0;
    };

    // zoom to center of image: (fx, fy) = (0.5, 0.5)
    const zoomInCentered = () => {
        setZoom(2);
        setPendingPan({ fx: 0.5, fy: 0.5 });
    };

    const handleZoomButtonClick = (event: MouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        if (zoom === 1) {
            zoomInCentered();
        } else {
            zoomOut();
        }
    };

    // if click the image at 1×, it zooms to 2× and pans so clicked spot is centered
    // if click again at 2×, it zooms back out.
    const handleImageClick = (event: MouseEvent<HTMLImageElement>) => {
        const scroller = scrollRef.current;
        if (!scroller) return; // if no scroll container, escape

        event.stopPropagation();

        if (zoom === 1) {
            const imgRect = event.currentTarget.getBoundingClientRect();
            const fx = (event.clientX - imgRect.left) / imgRect.width;
            const fy = (event.clientY - imgRect.top) / imgRect.height;

            setZoom(2);
            setPendingPan({ fx, fy });
        } else {
            zoomOut();
        }
    };

    // after zoom changes, pan so that the chosen point (fx, fy) is at the center
    useEffect(() => {
        if (zoom === 1 || !pendingPan) return;

        const scroller = scrollRef.current;
        if (!scroller) return;

        const scrollWidth = scroller.scrollWidth;
        const scrollHeight = scroller.scrollHeight;
        const clientWidth = scroller.clientWidth;
        const clientHeight = scroller.clientHeight;

        const maxLeft = Math.max(0, scrollWidth - clientWidth);
        const maxTop = Math.max(0, scrollHeight - clientHeight);

        const rawLeft = pendingPan.fx * scrollWidth - clientWidth / 2;
        const rawTop = pendingPan.fy * scrollHeight - clientHeight / 2;

        const targetLeft = Math.max(0, Math.min(rawLeft, maxLeft));
        const targetTop = Math.max(0, Math.min(rawTop, maxTop));

        scroller.scrollLeft = targetLeft;
        scroller.scrollTop = targetTop;

        // clear so this runs only once per zoom action
        setPendingPan(null);
    }, [zoom, pendingPan]);

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
                        <div className={styles.viewerImageArea}>
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
                                {zoom === 1 ? "1×" : "2×"}
                            </button>

                            {/* scrollable image area */}
                            <div
                                className={styles.viewerImageScroll}
                                ref={scrollRef}
                                style={{
                                    cursor: hoveringImage
                                        ? zoom === 1
                                            ? "zoom-in"
                                            : "zoom-out"
                                        : "default",
                                    // Center image at 1x; top-left anchor when zoomed
                                    justifyContent:
                                        zoom === 1 ? "center" : "flex-start",
                                    alignItems:
                                        zoom === 1 ? "center" : "flex-start",
                                }}
                            >
                                <img
                                    src={`http://localhost:3824/thumbnails/${selectedId}?size=2`}
                                    alt={`Preview of image ${selectedId}`}
                                    className={styles.viewerImage}
                                    draggable={false}
                                    onMouseEnter={() => setHoveringImage(true)}
                                    onMouseLeave={() => setHoveringImage(false)}
                                    onClick={handleImageClick}
                                    style={
                                        zoom === 1
                                            ? {}
                                            : {
                                                  width: `${zoom * 100}%`,
                                                  height: "auto",
                                                  maxWidth: "none",
                                                  maxHeight: "none",
                                              }
                                    }
                                />
                            </div>
                        </div>

                        <aside className={styles.viewerSidebar}>
                            <h2 className={styles.sidebarTitle}>
                                Image details
                            </h2>
                            <dl className={styles.sidebarList}>
                                {/* TODO: placeholders, swap for real metadata later */}
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
