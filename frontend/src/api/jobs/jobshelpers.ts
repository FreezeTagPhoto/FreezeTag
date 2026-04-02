export type JobSummary = {
    in_progress: number;
    complete: number;
    errors: number;
    uuid: string;
    status: string;
    title: string;
};

export type JobDetails = {
    // The job details endpoint is allowed to return anything in these arrays,
    // so for now we just use any and rely on knowledge of our input data to get what we expect

    /* eslint-disable @typescript-eslint/no-explicit-any */
    in_progress?: any[];
    completed?: any[];
    failed?: any[];
    /* eslint-enable @typescript-eslint/no-explicit-any */

    uuid: string;
    cancelled: boolean;
};
