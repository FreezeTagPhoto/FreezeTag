import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type UserCreateResult = Result<
    UserCreateResponse,
    { status: number; message: string }
>;
type UserCreateResponse = { created_at: number; id: number; username: string };

// The form should have a password and username field, both are strings
export default async function UserCreator(
    event: FormData,
): Promise<UserCreateResult> {
    return create_user_with_handler(
        ApiHandler<UserCreateResponse>(SERVER_ADDRESS + "createuser")(
            Method.POST,
        ),
        event,
    );
}

async function create_user_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<UserCreateResponse, RequestError>>,
    event: FormData,
): Promise<UserCreateResult> {
    const request_result = await handler(event);

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

export const testing_UserCreator = create_user_with_handler;
export type testing_UserCreateResponse = UserCreateResponse;
