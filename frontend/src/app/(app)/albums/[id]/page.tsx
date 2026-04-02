"use client";

import { use } from "react";
import AlbumDetailPage from "@/components/Albums/AlbumDetailPage/AlbumDetailPage";

type AlbumDetailRouteProps = {
    params: Promise<{ id: string }>;
};

export default function Page({ params }: AlbumDetailRouteProps) {
    const resolvedParams = use(params);
    const albumId = parseInt(resolvedParams.id, 10);

    if (isNaN(albumId)) {
        return <div>Invalid Album ID</div>;
    }

    return <AlbumDetailPage albumId={albumId} />;
}