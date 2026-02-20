import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { Perm } from "./permshelpers";

export type PermListResult = Result<
    Perm[],
    { status: number; message: string }
>;
type PermListResponse = Perm[];

export default async function PermsLister(): Promise<PermListResult> {
    return list_perms_with_handler(
        ApiHandler<PermListResponse>(SERVER_ADDRESS + "permissions/list")(
            Method.GET,
        ),
    );
}

async function list_perms_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PermListResponse, RequestError>>,
): Promise<PermListResult> {
    const request_result = await handler("");

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
