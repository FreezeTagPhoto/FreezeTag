import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PermsAddResult = Result<
    { message: string },
    { status: number; message: string }
>;
type PermsAddResponse = { message: string };

export default async function PermsAdder(
    user_id: number,
    permissions: string[],
): Promise<PermsAddResult> {
    return add_perms_with_handler(
        ApiHandler<PermsAddResponse>(
            SERVER_ADDRESS + "user/permissions/",
            false,
        )(Method.POST),
        user_id,
        permissions,
    );
}

async function add_perms_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PermsAddResponse, RequestError>>,
    user_id: number,
    permissions: string[],
): Promise<PermsAddResult> {
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
