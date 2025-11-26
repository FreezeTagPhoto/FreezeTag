import styles from "./GalleryImage.module.css";

export type GalleryImageProps = {
    id: number;
    onClick?: () => void;
    buttonRef?: (el: HTMLButtonElement | null) => void;
    onFocus?: () => void;
};

export default function GalleryImage({
    id,
    onClick,
    buttonRef,
    onFocus,
}: GalleryImageProps) {
    return (
        <button
            type="button"
            className={styles.image_container}
            onClick={onClick}
            ref={buttonRef}
            onFocus={onFocus}
        >
            <img
                src={`http://localhost:3824/thumbnails/${id}?size=1`}
                loading="lazy"
                alt={`A thumbnail of image ${id}`}
                decoding="async"
                className={styles.image}
            />
        </button>
    );
}
