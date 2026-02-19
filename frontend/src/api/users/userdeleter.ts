import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type UserDeleteResult = Result<
    { message: string },
    { status: number; message: string }
>;
type UserDeleteResponse = { message: string };

export default async function UserDeleter(
    user_id: number,
): Promise<UserDeleteResult> {
    return delete_user_with_handler(
        ApiHandler<UserDeleteResponse>(
            SERVER_ADDRESS + "users/",
            false,
        )(Method.DELETE),
        user_id,
    );
}

async function delete_user_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<UserDeleteResponse, RequestError>>,
    user_id: number,
): Promise<UserDeleteResult> {
    const query = `${user_id}`;
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
