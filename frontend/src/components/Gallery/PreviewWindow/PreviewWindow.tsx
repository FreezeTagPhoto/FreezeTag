"use client";

import {
    useCallback,
    useEffect,
    useLayoutEffect,
    useRef,
    useState,
    MouseEvent as ReactMouseEvent,
} from "react";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import styles from "./PreviewWindow.module.css";
import MetadataSidebar from "../MetadataSidebar/MetadataSidebar";
import {
    ChevronLeft,
    ChevronRight,
    Info,
    X,
    ZoomIn,
    ZoomOut,
} from "lucide-react";

type PendingPan = null | { fx: number; fy: number };
type BaseSize = null | { w: number; h: number };

export type PreviewWindowProps = {
    imageIds: number[];
    selectedId: number;
    onClose: () => void;

    onNavigate: (nextId: number, nextIndex: number) => void;
    onSearchTag?: (tag: string) => void;
    onDeleted?: (deletedId: number) => void;
};

export default function PreviewWindow({
    imageIds,
    selectedId,
    onClose,
    onNavigate,
    onSearchTag,
    onDeleted,
}: PreviewWindowProps) {
    const viewerRef = useRef<HTMLDivElement | null>(null);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const imgRef = useRef<HTMLImageElement | null>(null);

    const [zoom, setZoom] = useState<number>(1);
    const [hoveringImage, setHoveringImage] = useState(false);
    const [pendingPan, setPendingPan] = useState<PendingPan>(null);
    const [baseSize, setBaseSize] = useState<BaseSize>(null);
    const [sidebarOpen, setSidebarOpen] = useState(true);

    const [imgSrc, setImgSrc] = useState(
        `${SERVER_ADDRESS}/thumbnails/${selectedId}?size=1`,
    );
    const [highResReady, setHighResReady] = useState(false);
    const highResLoaderRef = useRef<HTMLImageElement | null>(null);

    const handleDeleted = useCallback(
        () => onDeleted?.(selectedId),
        [onDeleted, selectedId],
    );

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
        setBaseSize(null);

        const scroller = scrollRef.current;
        if (scroller) {
            scroller.scrollLeft = 0;
            scroller.scrollTop = 0;
        }
    }, [selectedId]);

    // when selectedId changes, show low-res immediately and preload high-res in background
    // the highResLoaderRef guard prevents stale loads from applying after navigation
    useEffect(() => {
        setImgSrc(`${SERVER_ADDRESS}/thumbnails/${selectedId}?size=1`);
        setHighResReady(false);

        const prev = highResLoaderRef.current;
        if (prev) {
            prev.onload = null;
            prev.onerror = null;
        }

        const loader = new Image();
        highResLoaderRef.current = loader;

        loader.onload = () => {
            if (highResLoaderRef.current !== loader) return;
            setImgSrc(`${SERVER_ADDRESS}/thumbnails/${selectedId}?size=2`);
            setHighResReady(true);
        };
        loader.onerror = () => {
            // high-res failed; stop blurring so the low-res is shown cleanly
            if (highResLoaderRef.current !== loader) return;
            setHighResReady(true);
        };
        loader.src = `${SERVER_ADDRESS}/thumbnails/${selectedId}?size=2`;

        // if size=2 is already in the browser cache, loader.complete will be true,
        // skip low-res phase
        if (loader.complete) {
            loader.onload = null;
            loader.onerror = null;
            highResLoaderRef.current = null;
            setImgSrc(`${SERVER_ADDRESS}/thumbnails/${selectedId}?size=2`);
            setHighResReady(true);
        }

        return () => {
            loader.onload = null;
            loader.onerror = null;
            if (highResLoaderRef.current === loader) {
                highResLoaderRef.current = null;
            }
        };
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
        return () => window.removeEventListener("keydown", handleKeyDown, true);
    }, [moveSelection, onClose]);

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

    const ensureBaseSize = useCallback(() => {
        const img = imgRef.current;
        const scroller = scrollRef.current;
        if (!img || !scroller) return;

        const nW = img.naturalWidth;
        const nH = img.naturalHeight;
        if (nW < 1 || nH < 1) return;

        const cW = scroller.clientWidth;
        const cH = scroller.clientHeight;
        if (cW < 1 || cH < 1) return;

        const imgAspect = nW / nH;
        const cAspect = cW / cH;
        const w = imgAspect > cAspect ? cW : cH * imgAspect;
        const h = imgAspect > cAspect ? cW / imgAspect : cH;

        setBaseSize({ w, h });
    }, []);

    useLayoutEffect(() => {
        const scroller = scrollRef.current;
        if (!scroller) return;

        const ro = new ResizeObserver(() => {
            if (zoom === 1) {
                requestAnimationFrame(() => ensureBaseSize());
            }
        });

        ro.observe(scroller);
        return () => ro.disconnect();
    }, [zoom, ensureBaseSize]);

    const handleImageLoad = () => {
        requestAnimationFrame(() => ensureBaseSize());
    };

    const handleImageClick = (event: ReactMouseEvent<HTMLImageElement>) => {
        const scroller = scrollRef.current;
        if (!scroller) return;

        event.stopPropagation();

        if (zoom === 1) {
            const imgRect = event.currentTarget.getBoundingClientRect();

            if (!baseSize) {
                const img = event.currentTarget;
                const nW = img.naturalWidth;
                const nH = img.naturalHeight;
                if (nW > 0 && nH > 0) {
                    const cW = scroller.clientWidth;
                    const cH = scroller.clientHeight;
                    const imgAspect = nW / nH;
                    const cAspect = cW / cH;
                    const w = imgAspect > cAspect ? cW : cH * imgAspect;
                    const h = imgAspect > cAspect ? cW / imgAspect : cH;
                    setBaseSize({ w, h });
                }
            }

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

    // preserve focal point so visible image content doesn't shift under sidebar column
    // (when toggling sidebar)
    const handleSidebarToggle = (e: ReactMouseEvent<HTMLButtonElement>) => {
        e.stopPropagation();

        if (zoom !== 1 && scrollRef.current) {
            const scroller = scrollRef.current;
            const fx =
                scroller.scrollWidth > 0
                    ? (scroller.scrollLeft + scroller.clientWidth / 2) /
                      scroller.scrollWidth
                    : 0.5;
            const fy =
                scroller.scrollHeight > 0
                    ? (scroller.scrollTop + scroller.clientHeight / 2) /
                      scroller.scrollHeight
                    : 0.5;
            setPendingPan({ fx, fy });
        }

        setSidebarOpen((v) => !v);
    };

    const zoomed = zoom !== 1;

    const cursor = hoveringImage
        ? zoomed
            ? "zoom-out"
            : "zoom-in"
        : "default";

    // if baseSize isn't ready for some reason, fall back to the old percentage approach
    const zoomedStyle: React.CSSProperties | undefined = !zoomed
        ? undefined
        : baseSize
          ? {
                width: `${baseSize.w * zoom}px`,
                height: `${baseSize.h * zoom}px`,
                maxWidth: "none",
                maxHeight: "none",
            }
          : {
                width: `${zoom * 100}%`,
                height: `${zoom * 100}%`,
                maxWidth: "none",
                maxHeight: "none",
            };

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div
                ref={viewerRef}
                className={`${styles.viewer} ${sidebarOpen ? "" : styles.viewerSidebarHidden}`}
                onClick={(e) => e.stopPropagation()}
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
                        className={styles.infoButton}
                        onClick={handleSidebarToggle}
                        aria-label={sidebarOpen ? "Hide info" : "Show info"}
                        title={sidebarOpen ? "Hide info" : "Show info"}
                    >
                        <Info className={styles.icon} />
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
                            cursor,
                            justifyContent:
                                zoom === 1 ? "center" : "flex-start",
                            alignItems: zoom === 1 ? "center" : "flex-start",
                        }}
                    >
                        <img
                            ref={imgRef}
                            src={imgSrc}
                            alt={`Preview of image ${selectedId}`}
                            className={styles.viewerImage}
                            draggable={false}
                            onMouseEnter={() => setHoveringImage(true)}
                            onMouseLeave={() => setHoveringImage(false)}
                            onClick={handleImageClick}
                            onLoad={handleImageLoad}
                            style={{
                                cursor,
                                filter: highResReady ? undefined : "blur(3px)",
                                // transition: "filter 10ms ease",
                                ...(zoomedStyle ?? {}),
                            }}
                        />
                    </div>
                </div>

                {sidebarOpen && (
                    <MetadataSidebar
                        selectedId={selectedId}
                        onSearchTag={onSearchTag}
                        viewerRef={viewerRef}
                        onDeleted={handleDeleted}
                    />
                )}
            </div>
        </div>
    );
}
