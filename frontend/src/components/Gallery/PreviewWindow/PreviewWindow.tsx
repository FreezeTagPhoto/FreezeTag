"use client";

import {
    useCallback,
    useEffect,
    useRef,
    useState,
    MouseEvent as ReactMouseEvent,
} from "react";
import styles from "./PreviewWindow.module.css";
import MetadataSidebar from "../MetadataSidebar/MetadataSidebar";

type PendingPan = null | { fx: number; fy: number };

export type PreviewWindowProps = {
    imageIds: number[];
    selectedId: number;
    onClose: () => void;

    /**
     * Called when user navigates via arrows / buttons.
     * Provide both ID + index so MainGallery can keep focusedIndex consistent.
     */
    onNavigate: (nextId: number, nextIndex: number) => void;

    onSearchTag?: (tag: string) => void;
};

export default function PreviewWindow({
    imageIds,
    selectedId,
    onClose,
    onNavigate,
    onSearchTag,
}: PreviewWindowProps) {
    // zoom: 1 = fit, 2 = zoomed
    const [zoom, setZoom] = useState<number>(1);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const [hoveringImage, setHoveringImage] = useState(false);

    // remember where to pan after zoom has been applied
    const [pendingPan, setPendingPan] = useState<PendingPan>(null);

    const moveSelection = useCallback(
        (direction: "next" | "prev") => {
            const currentIndex = imageIds.indexOf(selectedId);
            if (currentIndex === -1) return;

            const delta = direction === "next" ? 1 : -1;
            const nextIndex = currentIndex + delta;

            if (nextIndex < 0 || nextIndex >= imageIds.length) return;

            const nextId = imageIds[nextIndex];
            onNavigate(nextId, nextIndex);
        },
        [selectedId, imageIds, onNavigate],
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

    // keyboard: esc closes, arrows navigate
    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                onClose();
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
        return () =>
            window.removeEventListener("keydown", handleKeyDown, true);
    }, [moveSelection, onClose]);

    const handleBackdropClick = () => onClose();

    const stopPropagation = (e: ReactMouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
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

    const handleZoomButtonClick = (
        event: ReactMouseEvent<HTMLButtonElement>,
    ) => {
        event.stopPropagation();
        if (zoom === 1) {
            zoomInCentered();
        } else {
            zoomOut();
        }
    };

    // if click image at 1× -> zoom to 2× and pan so clicked spot is centered
    // if click again at 2× -> zoom out
    const handleImageClick = (event: ReactMouseEvent<HTMLImageElement>) => {
        const scroller = scrollRef.current;
        if (!scroller) return;

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

    // after zoom changes, pan so chosen point (fx, fy) is centered
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

        setPendingPan(null);
    }, [zoom, pendingPan]);

    return (
        <div className={styles.viewerBackdrop} onClick={handleBackdropClick}>
            <div className={styles.viewer} onClick={stopPropagation}>
                <div className={styles.viewerImageArea}>
                    {/* chevrons */}
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonLeft}`}
                        onClick={(e) => {
                            e.stopPropagation();
                            moveSelection("prev");
                        }}
                        aria-label="Previous image"
                    />
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonRight}`}
                        onClick={(e) => {
                            e.stopPropagation();
                            moveSelection("next");
                        }}
                        aria-label="Next image"
                    />

                    {/* close */}
                    <button
                        type="button"
                        className={styles.closeButton}
                        onClick={(e) => {
                            e.stopPropagation();
                            onClose();
                        }}
                        aria-label="Close"
                    />

                    {/* zoom toggle */}
                    <button
                        type="button"
                        className={styles.zoomButton}
                        onClick={handleZoomButtonClick}
                        aria-label={zoom === 1 ? "Zoom to 2x" : "Zoom to 1x"}
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
                            justifyContent: zoom === 1 ? "center" : "flex-start",
                            alignItems: zoom === 1 ? "center" : "flex-start",
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

                <MetadataSidebar
                    selectedId={selectedId}
                    onSearchTag={onSearchTag}
                />
            </div>
        </div>
    );
}