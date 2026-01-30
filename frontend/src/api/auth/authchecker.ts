import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";

type AuthCheckResponse = { user_id: string };

// The form should have a password and username field, both are strings
export default async function AuthChecker(): Promise<boolean> {
    return auth_check_with_handler(
        ApiHandler<AuthCheckResponse>(SERVER_ADDRESS + "login")(Method.GET),
    );
}

async function auth_check_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AuthCheckResponse, RequestError>>,
): Promise<boolean> {
    return true;
    return (await handler("")).ok;
}

export const testing_AuthChecker = auth_check_with_handler;
export type testing_AuthCheckResponse = AuthCheckResponse;
