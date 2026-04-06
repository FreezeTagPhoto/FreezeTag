"use client";

import { use } from "react";
import AlbumEditPage from "@/components/Albums/AlbumEditPage/AlbumEditPage";

interface PageProps {
    params: Promise<{ albumName: string }>;
}

export default function Page({ params }: PageProps) {
    const { albumName } = use(params);
    return <AlbumEditPage albumName={decodeURIComponent(albumName)} />;
}
