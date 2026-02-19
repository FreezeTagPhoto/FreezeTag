import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type UserGetResult = Result<
    { created_at: number; id: number; username: string },
    { status: number; message: string }
>;
type UserGetResponse = { created_at: number; id: number; username: string };

export default async function UserGetter(
    user_id: number,
): Promise<UserGetResult> {
    return get_user_with_handler(
        ApiHandler<UserGetResponse>(SERVER_ADDRESS + "users/")(Method.GET),
        user_id,
    );
}

async function get_user_with_handler(
    handler: (data: BodyInit) => Promise<Result<UserGetResponse, RequestError>>,
    user_id: number,
): Promise<UserGetResult> {
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
