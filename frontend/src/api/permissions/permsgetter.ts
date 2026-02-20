import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { Perm } from "./permshelpers";

export type PermsGetResult = Result<
    PermsGetResponse,
    { status: number; message: string }
>;
type PermsGetResponse = Perm[];

export default async function PermsGetter(
    user_id: number,
): Promise<PermsGetResult> {
    return get_perms_with_handler(
        ApiHandler<PermsGetResponse>(SERVER_ADDRESS + "user/permissions/")(
            Method.GET,
        ),
        user_id,
    );
}

async function get_perms_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PermsGetResponse, RequestError>>,
    user_id: number,
): Promise<PermsGetResult> {
    const request_result = await handler(`${user_id}`);

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
