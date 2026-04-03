import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import "leaflet/dist/leaflet.css";

import favicon from "@/icons/favicon.ico";
import InnerLayout from "./inner_layout";
import { NavigationGuardProvider } from "next-navigation-guard";

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
        <NavigationGuardProvider>
            <InnerLayout
                className={`${geistSans.variable} ${geistMono.variable}`}
            >
                {children}
            </InnerLayout>
        </NavigationGuardProvider>
    );
}
