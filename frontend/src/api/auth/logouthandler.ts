import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";

type LogoutResponse = { status: string };

// The form should have a password and username field, both are strings
export default async function LogoutHandler(): Promise<void> {
    return logout_with_handler(
        ApiHandler<LogoutResponse>(SERVER_ADDRESS + "logout")(Method.POST),
    );
}

async function logout_with_handler(
    handler: (data: BodyInit) => Promise<Result<LogoutResponse, RequestError>>,
): Promise<void> {
    await handler("");
}

export const testing_LogoutHandler = logout_with_handler;
export type testing_LogoutResponse = LogoutResponse;
