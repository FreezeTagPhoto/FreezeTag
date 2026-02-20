import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";
import { None, Option, Some } from "@/common/option";
import { PermedUser, UserHasPerm } from "@/api/permissions/permshelpers";

export type AuthCheckOption = Option<AuthCheckResponse>;

type AuthCheckResponse = PermedUser;

/**
 *
 * @param permission If you need to check auth for a specific permission, then pass it in here
 * @returns Whether the current user is authed or not
 */
export default async function AuthChecker(
    permission?: string,
): Promise<AuthCheckOption> {
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
): Promise<AuthCheckOption> {
    const result = await handler("");
    if (!result.ok) {
        return None();
    }
    if (permission && UserHasPerm(result.value, permission)) {
        return Some(result.value);
    }
    if (permission) {
        return None();
    }
    return Some(result.value);
}
