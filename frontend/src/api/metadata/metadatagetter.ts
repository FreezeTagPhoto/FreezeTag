"use server";

import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type ImageMetadata = {
    fileName: string | null;
    dateTaken: number | null;
    dateUploaded: number | null;
    cameraMake: string | null;
    cameraModel: string | null;
    latitude: number | null;
    longitude: number | null;
};

export type MetadataGetResult = Result<
    ImageMetadata,
    { status: number; message: string }
>;
type MetadataGetResponse = ImageMetadata;

/**
 * Fetch metadata for a single image (GET /metadata/id).
 */
export default async function MetadataGetter(
    image_id: number,
): Promise<MetadataGetResult> {
    return get_metadata_with_handler(
        ApiHandler<MetadataGetResponse>(SERVER_ADDRESS + "metadata")(
            Method.GET,
        ),
        image_id,
    );
}

async function get_metadata_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<MetadataGetResponse, RequestError>>,
    image_id: number,
): Promise<MetadataGetResult> {
    const result = await handler(`/${image_id}`);

    if (!result.ok) {
        const status = result.error.status_code;

        try {
            const body = (await result.error.response.json()) as {
                error?: string;
            };
            if (typeof body?.error === "string") {
                return Err({ status, message: body.error });
            }
        } catch {
            // ignore and fall back to text
        }

        return Err({
            status,
            message: await result.error.response.text(),
        });
    }

    return result;
}

export const testing_MetadataGetter = get_metadata_with_handler;
export type testing_MetadataGetResponse = MetadataGetResponse;
