export type JobSummary = {
    in_progress: number;
    complete: number;
    errors: number;
    uuid: string;
    status: string;
    title: string;
};

export type JobDetails = {
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
