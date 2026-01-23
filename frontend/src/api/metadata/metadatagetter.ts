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
        const resp = result.error.response;

        if (status === 400) {
            return Err({
                status,
                message: ((await resp.json()) as { error: string }).error,
            });
        } else {
            return Err({
                status,
                message: await resp.text(),
            });
        }
    }

    return result;
}

export const testing_MetadataGetter = get_metadata_with_handler;
export type testing_MetadataGetResponse = MetadataGetResponse;
