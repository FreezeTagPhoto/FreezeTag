import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type JobCancelResult = Result<
    JobCancelResponse,
    { status: number; message: string }
>;

type JobCancelResponse = {
    uuid: string;
};

export default async function JobsCanceller(
    event: string,
): Promise<JobCancelResult> {
    return job_query_with_handler(
        ApiHandler<JobCancelResponse>(
            SERVER_ADDRESS + "jobs/cancel/",
            false,
        )(Method.POST),
        event,
    );
}

async function job_query_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<JobCancelResponse, RequestError>>,
    job_code: string,
): Promise<JobCancelResult> {
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
