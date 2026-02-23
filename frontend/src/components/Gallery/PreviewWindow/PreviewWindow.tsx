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
    X,
    ZoomIn,
    ZoomOut,
    Download,
    Trash2,
    Loader2,
} from "lucide-react";
import FileDownloader from "@/api/files/filedownloader";
import FileDeleter from "@/api/files/filedeleter";

type PendingPan = null | { fx: number; fy: number };
type BaseSize = null | { w: number; h: number };

export type PreviewWindowProps = {
    imageIds: number[];
    selectedId: number;
    onClose: () => void;

    onNavigate: (nextId: number, nextIndex: number) => void;
    onSearchTag?: (tag: string) => void;
};

async function requestErrorToMessage(err: {
    status_code: number;
    response: Response;
}): Promise<string> {
    try {
        const text = await err.response.text();
        if (!text) return err.response.statusText || `HTTP ${err.status_code}`;

        try {
            const json = JSON.parse(text) as unknown;
            if (
                json &&
                typeof json === "object" &&
                "error" in json &&
                typeof (json as { error: unknown }).error === "string"
            ) {
                return (json as { error: string }).error;
            }
        } catch {
            // ignore
        }

        return text;
    } catch {
        return err.status_code === 0 ? "Network error" : `HTTP ${err.status_code}`;
    }
}

function triggerBrowserDownload(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
}

export default function PreviewWindow({
    imageIds,
    selectedId,
    onClose,
    onNavigate,
    onSearchTag,
}: PreviewWindowProps) {
    const viewerRef = useRef<HTMLDivElement | null>(null);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const imgRef = useRef<HTMLImageElement | null>(null);

    const [zoom, setZoom] = useState<number>(1);
    const [hoveringImage, setHoveringImage] = useState(false);
    const [pendingPan, setPendingPan] = useState<PendingPan>(null);
    const [baseSize, setBaseSize] = useState<BaseSize>(null);

    const [actionBusy, setActionBusy] = useState<null | "download" | "delete">(
        null,
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
        if (!img) return;

        const r = img.getBoundingClientRect();
        if (r.width > 1 && r.height > 1) {
            setBaseSize({ w: r.width, h: r.height });
        }
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

            if (!baseSize && imgRect.width > 1 && imgRect.height > 1) {
                setBaseSize({ w: imgRect.width, h: imgRect.height });
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

    const handleDownload = async (event: ReactMouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        if (actionBusy) return;

        setActionBusy("download");
        const result = await FileDownloader(selectedId)();

        if ("ok" in result && result.ok) {
            triggerBrowserDownload(result.value.blob, result.value.filename);
        } else {
            const msg = await requestErrorToMessage(
                (result as { ok: false; error: { status_code: number; response: Response } })
                    .error,
            );
            window.alert(`Download failed: ${msg}`);
        }

        setActionBusy(null);
    };

    const handleDelete = async (event: ReactMouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        if (actionBusy) return;

        const confirmed = window.confirm(
            "Delete this image? This cannot be undone.",
        );
        if (!confirmed) return;

        setActionBusy("delete");
        const result = await FileDeleter(selectedId)();

        if ("ok" in result && result.ok) {
            onClose();
        } else {
            const msg = await requestErrorToMessage(
                (result as { ok: false; error: { status_code: number; response: Response } })
                    .error,
            );
            window.alert(`Delete failed: ${msg}`);
        }

        setActionBusy(null);
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

    const downloadBusy = actionBusy === "download";
    const deleteBusy = actionBusy === "delete";

    return (
        <div className={styles.viewerBackdrop} onClick={() => onClose()}>
            <div
                ref={viewerRef}
                className={styles.viewer}
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

                    <div className={styles.actionBar}>
                        <button
                            type="button"
                            className={styles.actionButton}
                            onClick={handleZoomButtonClick}
                            aria-label={zoom === 1 ? "Zoom to 2x" : "Zoom to 1x"}
                            title={zoom === 1 ? "Zoom in" : "Zoom out"}
                            disabled={actionBusy !== null}
                        >
                            {zoom === 1 ? (
                                <ZoomIn className={styles.icon} />
                            ) : (
                                <ZoomOut className={styles.icon} />
                            )}
                        </button>

                        <button
                            type="button"
                            className={styles.actionButton}
                            onClick={handleDownload}
                            aria-label="Download image"
                            title="Download"
                            disabled={actionBusy !== null}
                        >
                            {downloadBusy ? (
                                <Loader2
                                    className={`${styles.icon} ${styles.spinning}`}
                                />
                            ) : (
                                <Download className={styles.icon} />
                            )}
                        </button>

                        <button
                            type="button"
                            className={`${styles.actionButton} ${styles.dangerButton}`}
                            onClick={handleDelete}
                            aria-label="Delete image"
                            title="Delete"
                            disabled={actionBusy !== null}
                        >
                            {deleteBusy ? (
                                <Loader2
                                    className={`${styles.icon} ${styles.spinning}`}
                                />
                            ) : (
                                <Trash2 className={styles.icon} />
                            )}
                        </button>
                    </div>

                    <div
                        className={styles.viewerImageScroll}
                        ref={scrollRef}
                        style={{
                            cursor,
                            justifyContent: zoom === 1 ? "center" : "flex-start",
                            alignItems: zoom === 1 ? "center" : "flex-start",
                        }}
                    >
                        <img
                            ref={imgRef}
                            src={`${SERVER_ADDRESS}/thumbnails/${selectedId}?size=2`}
                            alt={`Preview of image ${selectedId}`}
                            className={styles.viewerImage}
                            draggable={false}
                            onMouseEnter={() => setHoveringImage(true)}
                            onMouseLeave={() => setHoveringImage(false)}
                            onClick={handleImageClick}
                            onLoad={handleImageLoad}
                            style={{
                                cursor,
                                ...(zoomedStyle ?? {}),
                            }}
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
