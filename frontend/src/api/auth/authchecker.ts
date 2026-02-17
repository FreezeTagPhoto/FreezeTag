import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";

type AuthCheckResponse = { user_id: number; permissions: string[] };

/**
 *
 * @param permission If you need to check auth for a specific permission, then pass it in here
 * @returns Whether the current user is authed or not
 */
export default async function AuthChecker(
    permission?: string,
): Promise<boolean> {
    return auth_check_with_handler(
        ApiHandler<AuthCheckResponse>(SERVER_ADDRESS + "login")(Method.GET),
        permission,
    );
}

async function auth_check_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AuthCheckResponse, RequestError>>,
    permission?: string,
): Promise<boolean> {
    const result = await handler("");
    if (!result.ok) {
        return false;
    }
    if (permission && result.value.permissions.includes(permission)) {
        return true;
    }
    if (permission) {
        return false;
    }
    return true;
}
