import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { JobDetails } from "./jobshelpers";

export type JobsDetailResult = Result<
    JobDetails,
    { status: number; message: string }
>;

export default async function JobsDetailer(
    event: string,
): Promise<JobsDetailResult> {
    return job_detail_query_with_handler(
        ApiHandler<JobDetails>(SERVER_ADDRESS + "jobs/details/")(Method.GET),
        event,
    );
}

async function job_detail_query_with_handler(
    handler: (data: BodyInit) => Promise<Result<JobDetails, RequestError>>,
    job_code: string,
): Promise<JobsDetailResult> {
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

    return request_result;
}
