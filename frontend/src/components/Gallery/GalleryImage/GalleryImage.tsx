import SERVER_ADDRESS from "@/api/common/serveraddress";
import styles from "./GalleryImage.module.css";

export type GalleryImageProps = {
    id: number;
    onClick?: () => void;
    buttonRef?: (el: HTMLButtonElement | null) => void;
    onFocus?: () => void;
    selected?: boolean;
};

export default function GalleryImage({
    id,
    onClick,
    buttonRef,
    onFocus,
    selected,
}: GalleryImageProps) {
    return (
        <button
            type="button"
            className={
                selected
                    ? styles.image_container_fake_focus
                    : styles.image_container
            }
            onClick={onClick}
            ref={buttonRef}
            onFocus={onFocus}
        >
            <img
                src={`${SERVER_ADDRESS}/thumbnails/${id}?size=1`}
                loading="lazy"
                height={128}
                width={128}
                alt={`A thumbnail of image ${id}`}
                decoding="async"
                className={styles.image}
            />
        </button>
    );
}
