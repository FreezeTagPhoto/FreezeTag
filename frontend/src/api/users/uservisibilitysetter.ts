import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type UserVisibilitySetResult = Result<
    { message: string },
    { status: number; message: string }
>;

type UserVisibilitySetResponse = { message: string };

export default async function UserVisibilitySetter(
    userId: number,
    mode: number,
): Promise<UserVisibilitySetResult> {
    return set_visibility_with_handler(
        ApiHandler<UserVisibilitySetResponse>(
            SERVER_ADDRESS + `users/visibility/${userId}?mode=${mode}`,
        )(Method.POST),
    );
}

async function set_visibility_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<UserVisibilitySetResponse, RequestError>>,
): Promise<UserVisibilitySetResult> {
    const result = await handler("");

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        return Err({
            status,
            message: bodyText || "Failed to update user visibility.",
        });
    }

    return result;
}
