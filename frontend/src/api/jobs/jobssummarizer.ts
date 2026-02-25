import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { JobSummary } from "./jobshelpers";

export type JobSummaryResult = Result<
    JobSummary,
    { status: number; message: string }
>;

export default async function JobsSummarizer(
    event: string,
): Promise<JobSummaryResult> {
    return job_query_with_handler(
        ApiHandler<JobSummary>(SERVER_ADDRESS + "jobs/summary/")(Method.GET),
        event,
    );
}

async function job_query_with_handler(
    handler: (data: BodyInit) => Promise<Result<JobSummary, RequestError>>,
    job_code: string,
): Promise<JobSummaryResult> {
    const summary_request_result = await handler(job_code);

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
