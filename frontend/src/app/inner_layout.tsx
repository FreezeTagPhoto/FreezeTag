"use client";

import { ThemeGetter } from "@/themes/ThemeManager";

export default function InnerLayout({
    children,
    className,
}: {
    children: React.ReactNode;
    className?: string;
}) {
    return (
        <html lang="en" data-theme={ThemeGetter()}>
            <body className={className}>{children}</body>
        </html>
    );
}
