import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

import styles from "./page.module.css";

import Sidebar from "@/components/Sidebar/Sidebar";
import favicon from "@/icons/favicon.ico";

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
  description: "Free and open source self-hosted image tagging and management",
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
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        <div className={styles.shell}>
          <aside className={styles.nav}>
            <Sidebar />
          </aside>

          <div className={styles.content}>{children}</div>
        </div>
      </body>
    </html>
  );
}
