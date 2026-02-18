import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";
import JobsSummarizer from "./jobssummarizer";

// The outer result is Ok() if the request worked, and Err() if the request failed
// The inner result in Ok() is Ok() if the job is complete, and Err() if not. Err() is a fraction indicating progress
// The Ok(Ok(Map)) maps from image path to its ID on success, or to an error message on failure.
export type JobsResult = Result<
    Result<Map<string, Result<number, string>>, number>,
    { status: number; message: string }
>;

type JobResponse = {
    in_progress?: {
        name: string;
        status: string;
    }[];
    completed?: {
        filename: string;
        id: number;
    }[];
    failed?: {
        filename: string;
        reason: string;
    }[];
    uuid: string;
    cancelled: boolean;
};

export default async function JobsHandler(event: string): Promise<JobsResult> {
    return job_query_with_handler(
        ApiHandler<JobResponse>(SERVER_ADDRESS + "jobs/details/")(Method.GET),
        event,
    );
}

async function job_query_with_handler(
    handler: (data: BodyInit) => Promise<Result<JobResponse, RequestError>>,
    job_code: string,
): Promise<JobsResult> {
    const summary_request_result = await JobsSummarizer(job_code);

    if (!summary_request_result.ok) {
        return summary_request_result;
    }

    const summary_job_response = summary_request_result.value;

    if (summary_job_response.in_progress != 0) {
        return Ok(
            Err(
                (summary_job_response.complete + summary_job_response.errors) /
                    (summary_job_response.complete +
                        summary_job_response.errors +
                        summary_job_response.in_progress),
            ),
        );
    }

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

    const completed = job_response.completed ?? [];
    const failed = job_response.failed ?? [];

    const image_map = new Map();
    for (const result of completed) {
        image_map.set(result.filename, Ok(result.id));
    }

    for (const result of failed) {
        image_map.set(result.filename, Err(result.reason));
    }

    return Ok(Ok(image_map));
}
