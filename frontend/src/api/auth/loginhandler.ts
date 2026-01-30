import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";
import { Option, Some, None } from "@/common/option";

// Returns the error response if login didn't succeed,
// otherwise the token is put into cookies and other code doesn't need access to it
export type LoginError = Option<{ status: number; message: string }>;
type LoginResponse = { token: string };

// The form should have a password and username field, both are strings
export default async function LoginHandler(
    event: FormData,
): Promise<LoginError> {
    return login_with_handler(
        ApiHandler<LoginResponse>(SERVER_ADDRESS + "login")(Method.POST),
        event,
    );
}

async function login_with_handler(
    handler: (data: BodyInit) => Promise<Result<LoginResponse, RequestError>>,
    event: FormData,
): Promise<LoginError> {
    const request_result = await handler(event);

    if (!request_result.ok) {
        const status = request_result.error.status_code;
        if (status == 400)
            return Some({
                status,
                message: (
                    (await request_result.error.response.json()) as {
                        error: string;
                    }
                ).error,
            });
        else
            return Some({
                status,
                message: await request_result.error.response.text(),
            });
    }

    return None();
}

export const testing_LoginHandler = login_with_handler;
export type testing_LoginResponse = LoginResponse;
