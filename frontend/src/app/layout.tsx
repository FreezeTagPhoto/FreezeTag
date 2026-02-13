import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

import favicon from "@/icons/favicon.ico";
import InnerLayout from "./inner_layout";

const geistSans = Geist({
    variable: "--font-geist-sans",
    subsets: ["latin"],
});

const geistMono = Geist_Mono({
    variable: "--font-geist-mono",
    subsets: ["latin"],
});

export const metadata: Metadata = {
    title: "FreezeTag",
    description:
        "Free and open source self-hosted image tagging and management",
    icons: {
        icon: favicon.src,
    },
};

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <InnerLayout
            children={children}
            className={`${geistSans.variable} ${geistMono.variable}`}
        ></InnerLayout>
    );
}
