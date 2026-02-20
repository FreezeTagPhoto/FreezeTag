import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PermsRevokeResult = Result<
    { message: string },
    { status: number; message: string }
>;
type PermsRevokeResponse = { message: string };

export default async function PermsRevoker(
    user_id: number,
    permissions: string[],
): Promise<PermsRevokeResult> {
    return revoke_perms_with_handler(
        ApiHandler<PermsRevokeResponse>(
            SERVER_ADDRESS + "users/permissions/",
            false,
        )(Method.DELETE),
        user_id,
        permissions,
    );
}

async function revoke_perms_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PermsRevokeResponse, RequestError>>,
    user_id: number,
    permissions: string[],
): Promise<PermsRevokeResult> {
    const query = `${user_id}?permission=${permissions.join()}`;
    const request_result = await handler(query);

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
