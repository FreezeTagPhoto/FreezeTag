"use client";

import {
    ThemeGetter,
    ThemeTypeGetter,
    THEME_STORAGE_KEY,
    ApplyTheme,
} from "@/themes/ThemeManager";
import {
    UNIT_STORAGE_KEY,
    UnitsGetter,
    type UnitSystem,
} from "@/common/units/UnitManager";
import { useEffect, useState } from "react";

const THEME_CHANGED_EVENT = "freezetag:theme-changed";
const UNITS_CHANGED_EVENT = "freezetag:units-changed";

export default function InnerLayout({
    children,
    className,
}: {
    children: React.ReactNode;
    className?: string;
}) {
    const [theme, setTheme] = useState("");
    const [themeType, setThemeType] = useState("");
    const [units, setUnits] = useState<UnitSystem>("metric");

    useEffect(() => {
        const refreshTheme = () => {
            const t = ThemeGetter();
            setTheme(t);
            setThemeType(ThemeTypeGetter());
            ApplyTheme(t);
        };

        const refreshUnits = () => {
            setUnits(UnitsGetter());
        };

        refreshTheme();
        refreshUnits();

        const onStorage = (e: StorageEvent) => {
            if (e.key === THEME_STORAGE_KEY) refreshTheme();
            if (e.key === UNIT_STORAGE_KEY) refreshUnits();
        };

        window.addEventListener(THEME_CHANGED_EVENT, refreshTheme);
        window.addEventListener(UNITS_CHANGED_EVENT, refreshUnits);
        window.addEventListener("storage", onStorage);

        return () => {
            window.removeEventListener(THEME_CHANGED_EVENT, refreshTheme);
            window.removeEventListener(UNITS_CHANGED_EVENT, refreshUnits);
            window.removeEventListener("storage", onStorage);
        };
    }, []);

    return (
        <html
            lang="en"
            data-theme={theme}
            data-theme-type={themeType}
            data-units={units}
        >
            <body className={className}>{children}</body>
        </html>
    );
}
