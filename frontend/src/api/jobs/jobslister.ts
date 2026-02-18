import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type JobListResult = Result<
    JobListResponse,
    { status: number; message: string }
>;

type JobListResponse = {
    in_progress: number;
    complete: number;
    errors: number;
    uuid: string;
    status: string;
    title: string;
}[];

export default async function JobsLister(): Promise<JobListResult> {
    return job_query_with_handler(
        ApiHandler<JobListResponse>(SERVER_ADDRESS + "jobs/list")(Method.GET),
    );
}

async function job_query_with_handler(
    handler: (data: BodyInit) => Promise<Result<JobListResponse, RequestError>>,
): Promise<JobListResult> {
    const summary_request_result = await handler("");

    if (!summary_request_result.ok) {
        const status = summary_request_result.error.status_code;
        if (status == 400)
            return Err({
                status,
                message: (
                    (await summary_request_result.error.response.json()) as {
                        error: string;
                    }
                ).error,
            });
        else
            return Err({
                status,
                message: await summary_request_result.error.response.text(),
            });
    }

    return summary_request_result;
}
