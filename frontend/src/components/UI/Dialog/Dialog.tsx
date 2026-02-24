"use client";

import { useEffect, type ReactNode } from "react";
import styles from "./Dialog.module.css";
import { X } from "lucide-react";

export type DialogSize = "sm" | "md" | "lg";

export type DialogProps = {
    open: boolean;
    onClose: () => void;

    title: ReactNode;
    icon?: ReactNode;

    children?: ReactNode;
    actions?: ReactNode;

    ariaLabel?: string;

    size?: DialogSize;

    disableClose?: boolean;
    closeOnOverlayClick?: boolean;
    closeOnEscape?: boolean;

    showCloseButton?: boolean;
};

export default function Dialog({
    open,
    onClose,
    title,
    icon,
    children,
    actions,
    ariaLabel,
    size = "md",
    disableClose = false,
    closeOnOverlayClick = true,
    closeOnEscape = true,
    showCloseButton = true,
}: DialogProps) {
    useEffect(() => {
        if (!open) return;
        if (!closeOnEscape) return;

        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key !== "Escape") return;
            if (disableClose) return;
            onClose();
        };

        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, [open, closeOnEscape, disableClose, onClose]);

    if (!open) return null;

    return (
        <div
            className={styles.overlay}
            role="dialog"
            aria-modal="true"
            aria-label={ariaLabel}
            onMouseDown={(e) => {
                if (!closeOnOverlayClick) return;
                if (disableClose) return;
                if (e.target === e.currentTarget) onClose();
            }}
        >
            <div className={styles.modal} data-size={size}>
                <div className={styles.header}>
                    <div className={styles.titleRow}>
                        {icon && <div className={styles.iconSlot}>{icon}</div>}
                        <h2 className={styles.title}>{title}</h2>
                    </div>

                    {showCloseButton && (
                        <button
                            className={styles.close}
                            onClick={() => {
                                if (disableClose) return;
                                onClose();
                            }}
                            aria-label="Close"
                            title="Close"
                            disabled={disableClose}
                        >
                            <X className={styles.icon} />
                        </button>
                    )}
                </div>

                {children && <div className={styles.body}>{children}</div>}

                {actions && <div className={styles.footer}>{actions}</div>}
            </div>
        </div>
    );
}