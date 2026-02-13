"use client";

import { ThemeGetter } from "@/themes/ThemeManager";
import { useEffect, useState } from "react";

export default function InnerLayout({
    children,
    className,
}: {
    children: React.ReactNode;
    className?: string;
}) {
    const [theme, setTheme] = useState("");
    useEffect(() => {
        setTheme(ThemeGetter());
    }, []);
    return (
        <html lang="en" data-theme={theme}>
            <body className={className}>{children}</body>
        </html>
    );
}
