"use client";

import { useEffect, useRef } from "react";
import L from "leaflet";
import { ExternalLink } from "lucide-react";
import styles from "./LocationMap.module.css";

type LocationMapProps = {
    lat: number;
    lon: number;
};

const OSM_HREF = (lat: number, lon: number) =>
    `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}&zoom=13`;

export default function LocationMap({ lat, lon }: LocationMapProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const mapRef = useRef<L.Map | null>(null);

    useEffect(() => {
        if (!containerRef.current) return;

        mapRef.current?.remove();

        const map = L.map(containerRef.current, {
            center: [lat, lon],
            zoom: 13,
            zoomControl: false,
            attributionControl: false,
            dragging: false,
            scrollWheelZoom: false,
            doubleClickZoom: false,
            boxZoom: false,
            keyboard: false,
        });

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            maxZoom: 19,
        }).addTo(map);

        // read red accent color from CSS
        const accentColor =
            getComputedStyle(document.documentElement)
                .getPropertyValue("--red")
                .trim() || "#b45000";

        const pinSvg = `<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="${accentColor}" stroke="${accentColor}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 13-8 13s-8-7-8-13a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3" fill="white" stroke="none"/></svg>`;

        const icon = L.divIcon({
            html: pinSvg,
            className: "",
            iconSize: [28, 28],
            iconAnchor: [14, 28],
        });

        L.marker([lat, lon], { icon }).addTo(map);

        mapRef.current = map;

        return () => {
            mapRef.current?.remove();
            mapRef.current = null;
        };
    }, [lat, lon]);

    const href = OSM_HREF(lat, lon);

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
