"use client";

import { ThemeGetter, ThemeTypeGetter } from "@/themes/ThemeManager";
import { useEffect, useState } from "react";

export default function InnerLayout({
    children,
    className,
}: {
    children: React.ReactNode;
    className?: string;
}) {
    const [theme, setTheme] = useState("");
    const [themeType, setThemeType] = useState("");
    useEffect(() => {
        setTheme(ThemeGetter());
        setThemeType(ThemeTypeGetter());
    }, []);
    return (
        <html lang="en" data-theme={theme} data-theme-type={themeType}>
            <body className={className}>{children}</body>
        </html>
    );
}
