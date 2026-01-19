"use server";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

// The outer result is Ok() if the request worked, and Err() if the request failed
// The inner result in Ok() is Ok() if the job is complete, and Err() if not. Err() is a fraction indicating progress
// The Ok(Ok(Map)) maps from image path to its ID on success, or to an error message on failure.
export type JobsResult = Result<
    Result<Map<string, Result<number, string>>, number>,
    { status: number; message: string }
>;
type JobResponse = {
    in_progress: {
        name: string;
        status: string;
    }[];
    results: {
        error:
            | {
                  filename: string;
                  reason: string;
              }
            | undefined;
        success:
            | {
                  filename: string;
                  id: number;
              }
            | undefined;
    }[];
    uuid: string;
};

export default async function JobsHandler(event: string): Promise<JobsResult> {
    return job_query_with_handler(
        ApiHandler<JobResponse>(SERVER_ADDRESS + "jobquery/")(Method.GET),
        event,
    );
}

async function job_query_with_handler(
    handler: (data: BodyInit) => Promise<Result<JobResponse, RequestError>>,
    job_code: string,
): Promise<JobsResult> {
    const request_result = await handler(job_code);

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

    const job_response = request_result.value;
    const count_in_progress = job_response.in_progress.length;
    const count_done = job_response.results ? job_response.results.length : 0;

    if (count_in_progress != 0) {
        return Ok(Err(count_done / (count_done + count_in_progress)));
    }

    const body = job_response.results;

    const image_map = new Map();
    for (const result of body) {
        if (result.error) {
            image_map.set(result.error.filename, Err(result.error.reason));
        } else {
            if (!result.success) {
                console.error(
                    `Image didn't have any data in the return! Job code: ${job_code}`,
                );
                continue;
            }
            image_map.set(result.success.filename, Ok(result.success.id));
        }
    }

    return Ok(Ok(image_map));
}

export const testing_JobsHandler = job_query_with_handler;
export type testing_JobResponse = JobResponse;
