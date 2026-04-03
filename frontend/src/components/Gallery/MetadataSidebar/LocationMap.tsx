"use client";

import { useEffect, useRef, useState } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { ExternalLink, MapPin } from "lucide-react";
import styles from "./LocationMap.module.css";

type LocationMapProps = {
    lat: number;
    lon: number;
};

interface LeafletLayer {
    addTo(map: LeafletMap): void;
}

interface LeafletMap {
    remove(): void;
}

interface LeafletDivIconOptions {
    html: string;
    className: string;
    iconSize: [number, number];
    iconAnchor: [number, number];
}

interface LeafletStatic {
    map(el: HTMLElement, options: Record<string, unknown>): LeafletMap;
    tileLayer(url: string, options: Record<string, unknown>): LeafletLayer;
    marker(latlng: [number, number], options?: Record<string, unknown>): LeafletLayer;
    divIcon(options: LeafletDivIconOptions): Record<string, unknown>;
}

declare global {
    interface Window {
        L: LeafletStatic;
    }
}

const LEAFLET_VERSION = "1.9.4";
const LEAFLET_CSS = `https://unpkg.com/leaflet@${LEAFLET_VERSION}/dist/leaflet.css`;
const LEAFLET_JS = `https://unpkg.com/leaflet@${LEAFLET_VERSION}/dist/leaflet.js`;

// Module-level singleton — Leaflet is loaded at most once per page
let leafletState: "idle" | "loading" | "ready" | "error" = "idle";
const leafletCallbacks: Array<(ok: boolean) => void> = [];

function loadLeaflet(onDone: (ok: boolean) => void): void {
    if (leafletState === "ready") {
        onDone(true);
        return;
    }
    if (leafletState === "error") {
        onDone(false);
        return;
    }

    leafletCallbacks.push(onDone);

    if (leafletState === "loading") return;
    leafletState = "loading";

    if (!document.querySelector(`link[href="${LEAFLET_CSS}"]`)) {
        const link = document.createElement("link");
        link.rel = "stylesheet";
        link.href = LEAFLET_CSS;
        document.head.appendChild(link);
    }

    const script = document.createElement("script");
    script.src = LEAFLET_JS;
    script.onload = () => {
        leafletState = "ready";
        leafletCallbacks.splice(0).forEach((cb) => cb(true));
    };
    script.onerror = () => {
        leafletState = "error";
        leafletCallbacks.splice(0).forEach((cb) => cb(false));
    };
    document.head.appendChild(script);
}

const OSM_HREF = (lat: number, lon: number) =>
    `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}&zoom=13`;

export default function LocationMap({ lat, lon }: LocationMapProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const mapRef = useRef<LeafletMap | null>(null);
    const [loadError, setLoadError] = useState(false);

    useEffect(() => {
        let cancelled = false;

        loadLeaflet((ok) => {
            if (cancelled || !containerRef.current) return;
            if (!ok) {
                setLoadError(true);
                return;
            }

            if (mapRef.current) {
                mapRef.current.remove();
                mapRef.current = null;
            }

            const L = window.L;

            const map = L.map(containerRef.current!, {
                center: [lat, lon],
                zoom: 13,
                zoomControl: false,
                attributionControl: false,
                dragging: false,
                scrollWheelZoom: false,
                doubleClickZoom: false,
                boxZoom: false,
                keyboard: false,
                tap: false,
            });

            L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
                maxZoom: 19,
            }).addTo(map);

            // read red accent color from CSS
            const accentColor =
                getComputedStyle(document.documentElement)
                    .getPropertyValue("--red")
                    .trim() || "#b45000";

            // lucide-style MapPin SVG, filled with the accent color
            const pinSvg = `<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="${accentColor}" stroke="${accentColor}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 13-8 13s-8-7-8-13a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3" fill="white" stroke="none"/></svg>`;

            const icon = L.divIcon({
                html: pinSvg,
                className: "",
                iconSize: [28, 28],
                iconAnchor: [14, 28],
            });

            L.marker([lat, lon], { icon }).addTo(map);

            mapRef.current = map;
        });

        return () => {
            cancelled = true;
            mapRef.current?.remove();
            mapRef.current = null;
        };
    }, [lat, lon]);

    const href = OSM_HREF(lat, lon);

    if (loadError) {
        return (
            <a
                className={styles.osmLink}
                href={href}
                target="_blank"
                rel="noopener noreferrer"
            >
                View in OpenStreetMap
                <ExternalLink className={styles.osmLinkIcon} />
            </a>
        );
    }

    return (
        <div className={styles.mapWrap}>
            <div ref={containerRef} className={styles.mapContainer} />
            <a
                className={styles.osmLink}
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                onClick={(e) => e.stopPropagation()}
            >
                View in OpenStreetMap
                <ExternalLink className={styles.osmLinkIcon} />
            </a>
        </div>
    );
}
