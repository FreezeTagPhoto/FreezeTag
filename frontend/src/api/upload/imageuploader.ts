"use client";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

// Returns the UUID of the Job
export type UploadResult = Result<string, { status: number; message: string }>;
type UploadResponse = string;

export default async function ImageUploader(
    event: FormData,
): Promise<UploadResult> {
    return image_upload_with_handler(
        ApiHandler<UploadResponse>((await SERVER_ADDRESS()) + "upload")(
            Method.POST,
        ),
        event,
    );
}

async function image_upload_with_handler(
    handler: (data: BodyInit) => Promise<Result<UploadResponse, RequestError>>,
    event: FormData,
): Promise<UploadResult> {
    const request_result = await handler(event);

    if (!request_result.ok) {
        const status = request_result.error.status_code;
        if (status == 400)
            return Err({
                status,
                message: (
                    (await request_result.error.response.json()) as {
                        error: string;
                    }
                ).error,
            });
        else
            return Err({
                status,
                message: await request_result.error.response.text(),
            });
    }

    return Ok(request_result.value);
}

export const testing_ImageUploader = image_upload_with_handler;
export type testing_UploadResponse = UploadResponse;
