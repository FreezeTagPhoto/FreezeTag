"use client";

import SERVER_ADDRESS from "@/api/common/serveraddress";
import { type RequestError } from "@/api/common/apihandler";
import { type Result, Ok, Err } from "@/common/result";

export type FileDownloadSuccess = {
    blob: Blob;
    filename: string;
    contentType: string;
};

function parseContentDispositionFilename(header: string | null): string | null {
    if (!header) return null;

    const utf8 = header.match(/filename\*\s*=\s*UTF-8''([^;]+)/i);
    if (utf8?.[1]) {
        return decodeURIComponent(utf8[1].trim().replace(/^"+|"+$/g, ""));
    }

    const basic = header.match(/filename\s*=\s*("?)([^";]+)\1/i);
    if (basic?.[2]) return basic[2].trim();

    return null;
}

export default function FileDownloader(id: number) {
    return async (): Promise<Result<FileDownloadSuccess, RequestError>> => {
        let response: Response;

        try {
            response = await fetch(`${SERVER_ADDRESS}/file/download/${id}`, {
                method: "GET",
                headers: {
                    accept: "application/octet-stream",
                },
            });
        } catch (error) {
            return Err({
                status_code: 0,
                response: new Response(`Catastrophic Error: ${error}`),
            });
        }

        if (!response.ok) {
            return Err({ status_code: response.status, response });
        }

        const blob = await response.blob();
        const contentType =
            response.headers.get("content-type") ??
            blob.type ??
            "application/octet-stream";

        const filename =
            parseContentDispositionFilename(
                response.headers.get("content-disposition"),
            ) ?? `image-${id}`;

        return Ok({ blob, filename, contentType });
    };
}
