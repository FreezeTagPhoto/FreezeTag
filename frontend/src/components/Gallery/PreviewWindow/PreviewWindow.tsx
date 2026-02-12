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
import { ChevronLeft, ChevronRight, X, ZoomIn, ZoomOut } from "lucide-react";

type PendingPan = null | { fx: number; fy: number };

export type PreviewWindowProps = {
    imageIds: number[];
    selectedId: number;
    onClose: () => void;

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
    const viewerRef = useRef<HTMLDivElement | null>(null);

    const [zoom, setZoom] = useState<number>(1);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const [hoveringImage, setHoveringImage] = useState(false);
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
        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                onClose();
                return;
            }

            if (event.key !== "ArrowRight" && event.key !== "ArrowLeft") return;

            event.preventDefault();
            event.stopPropagation();

            if (event.key === "ArrowRight") moveSelection("next");
            else moveSelection("prev");
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

    const zoomInCentered = () => {
        setZoom(2);
        setPendingPan({ fx: 0.5, fy: 0.5 });
    };

    const handleZoomButtonClick = (
        event: ReactMouseEvent<HTMLButtonElement>,
    ) => {
        event.stopPropagation();
        if (zoom === 1) zoomInCentered();
        else zoomOut();
    };

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

        scroller.scrollLeft = Math.max(0, Math.min(rawLeft, maxLeft));
        scroller.scrollTop = Math.max(0, Math.min(rawTop, maxTop));

        setPendingPan(null);
    }, [zoom, pendingPan]);

    return (
        <div className={styles.viewerBackdrop} onClick={handleBackdropClick}>
            <div
                ref={viewerRef}
                className={styles.viewer}
                onClick={stopPropagation}
            >
                <div className={styles.viewerImageArea}>
                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonLeft}`}
                        onClick={(e) => {
                            e.stopPropagation();
                            moveSelection("prev");
                        }}
                        aria-label="Previous image"
                        title="Previous"
                    >
                        <ChevronLeft className={styles.iconLg} />
                    </button>

                    <button
                        type="button"
                        className={`${styles.navButton} ${styles.navButtonRight}`}
                        onClick={(e) => {
                            e.stopPropagation();
                            moveSelection("next");
                        }}
                        aria-label="Next image"
                        title="Next"
                    >
                        <ChevronRight className={styles.iconLg} />
                    </button>

                    <button
                        type="button"
                        className={styles.closeButton}
                        onClick={(e) => {
                            e.stopPropagation();
                            onClose();
                        }}
                        aria-label="Close"
                        title="Close"
                    >
                        <X className={styles.iconLg} />
                    </button>

                    <button
                        type="button"
                        className={styles.zoomButton}
                        onClick={handleZoomButtonClick}
                        aria-label={zoom === 1 ? "Zoom to 2x" : "Zoom to 1x"}
                        title={zoom === 1 ? "Zoom in" : "Zoom out"}
                    >
                        {zoom === 1 ? (
                            <ZoomIn className={styles.icon} />
                        ) : (
                            <ZoomOut className={styles.icon} />
                        )}
                    </button>

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
                    viewerRef={viewerRef}
                />
            </div>
        </div>
    );
}