"use client";

import {
    useEffect,
    useState,
    useRef,
    MouseEvent,
    KeyboardEvent as ReactKeyboardEvent,
    useCallback,
} from "react";
import styles from "./MainGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";
import MetadataGetter, { ImageMetadata } from "@/api/metadata/metadatagetter";

export type GalleryProps = {
    image_ids: number[];
};

// point (fx, fy) on image expressed as fraction of width/height (after zoom)
type PendingPan = null | { fx: number; fy: number };

// if int64 timestamp is huge, treat as ms, otherwise seconds
function toDate(ts: number): Date {
    return new Date(ts > 1e12 ? ts : ts * 1000);
}

function formatDate(ts: number | null): string {
    if (ts === null) return "—";
    const d = toDate(ts);
    return new Intl.DateTimeFormat(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
    }).format(d);
}

function formatLocation(lat: number | null, lon: number | null): string {
    if (lat === null || lon === null) return "—";
    return `${lat.toFixed(5)}, ${lon.toFixed(5)}`;
}

function formatCamera(make: string | null, model: string | null): string {
    const parts = [make, model].filter((x) => x && x.trim().length > 0) as string[];
    return parts.length ? parts.join(" ") : "—";
}

export default function MainGallery({ image_ids }: GalleryProps) {
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

    // metadata state
    const [metadataById, setMetadataById] = useState<Record<number, ImageMetadata>>({});
    const [metadataLoading, setMetadataLoading] = useState(false);
    const [metadataError, setMetadataError] = useState<string | null>(null);

    const currentMetadata: ImageMetadata | null =
        selectedId !== null ? metadataById[selectedId] ?? null : null;

    // fetch metadata whenever selectedId changes
    useEffect(() => {
        if (selectedId === null) return;

        // already cached
        if (metadataById[selectedId]) {
            setMetadataError(null);
            setMetadataLoading(false);
            return;
        }

        let cancelled = false;

        (async () => {
            setMetadataLoading(true);
            setMetadataError(null);

            const res = await MetadataGetter(selectedId);

            if (cancelled) return;

            if (!res.ok) {
                setMetadataError(res.error.message);
                setMetadataLoading(false);
                return;
            }

            setMetadataById((prev) => ({ ...prev, [selectedId]: res.value }));
            setMetadataLoading(false);
        })();

        return () => {
            cancelled = true;
        };
    }, [selectedId, metadataById]);

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
                            {metadataLoading && (
                                <p className={styles.sidebarHint}>Loading metadata…</p>
                            )}
                            {metadataError && (
                                <p className={styles.sidebarError}>
                                    Failed to load metadata: {metadataError}
                                </p>
                            )}
                            <dl className={styles.sidebarList}>
                                <div>
                                    <dt>Filename</dt>
                                    <dd>{currentMetadata?.fileName ?? "—"}</dd>
                                </div>
                                <div>
                                    <dt>Date taken</dt>
                                    <dd>
                                        {currentMetadata
                                            ? formatDate(currentMetadata.dateTaken)
                                            : "—"}
                                    </dd>
                                </div>

                                <div>
                                    <dt>Date uploaded</dt>
                                    <dd>
                                        {currentMetadata
                                            ? formatDate(currentMetadata.dateUploaded)
                                            : "—"}
                                    </dd>
                                </div>
                                <div>
                                    <dt>Location</dt>
                                    <dd>
                                        {currentMetadata
                                            ? formatLocation(
                                                  currentMetadata.latitude,
                                                  currentMetadata.longitude,
                                              )
                                            : "—"}
                                    </dd>
                                </div>
                                <div>
                                    <dt>Camera</dt>
                                    <dd>
                                        {currentMetadata
                                            ? formatCamera(
                                                  currentMetadata.cameraMake,
                                                  currentMetadata.cameraModel,
                                              )
                                            : "—"}
                                    </dd>
                                </div>
                            </dl>
                        </aside>
                    </div>
                </div>
            )}
        </>
    );
}
