import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type User = {
    created_at: number;
    id: number;
    username: string;
    visibility_mode?: number;
};
export type UserListResult = Result<
    User[],
    { status: number; message: string }
>;
type UserListResponse = User[];

export default async function UserLister(): Promise<UserListResult> {
    return list_user_with_handler(
        ApiHandler<UserListResponse>(SERVER_ADDRESS + "users/all")(Method.GET),
    );
}

async function list_user_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<UserListResponse, RequestError>>,
): Promise<UserListResult> {
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
