"use server";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

// Maps from image path to its ID on success, or to an error message on failure.
export type UploadResult = Result<
    Map<string, Result<number, string>>,
    { status: number; message: string }
>;
type UploadResponse = string;
type JobResponse = {
    in_progress: {
        name: string;
        status: string;
    }[];
    results: {
        error: {
            filename: string;
            reason: string;
        };
        success: {
            filename: string;
            id: 0;
        };
    }[];
    uuid: string;
};

export default async function ImageUploader(
    event: FormData,
): Promise<UploadResult> {
    return image_upload_with_handler(
        ApiHandler<UploadResponse>(SERVER_ADDRESS + "upload")(Method.POST),
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

    const job_id = request_result.value;

    const job_handler = ApiHandler<JobResponse>(SERVER_ADDRESS + "jobquery/")(
        Method.GET,
    );
    let job_query;
    while (true) {
        job_query = await job_handler(job_id);
        if (!job_query.ok) {
            return Err({ status: 7000, message: "" });
        }
        if (!job_query.value.in_progress.length) {
            break;
        }
        await new Promise((resolve) => setTimeout(resolve, 1000));
    }

    const body = job_query.value.results;

    const image_map = new Map();
    for (const result of body) {
        if (result.error) {
            image_map.set(result.error.filename, Err(result.error.reason));
        } else {
            image_map.set(result.success.filename, Ok(result.success.id));
        }
    }

    return Ok(image_map);
}

export const testing_ImageUploader = image_upload_with_handler;
export type testing_UploadResponse = UploadResponse;
