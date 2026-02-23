import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PasswordChangeResult = Result<
    PasswordChangeResponse,
    { status: number; message: string }
>;

type PasswordChangeResponse = { message: string };

export default async function PasswordChanger(
    event: FormData,
): Promise<PasswordChangeResult> {
    return change_password_with_handler(
        ApiHandler<PasswordChangeResponse>(SERVER_ADDRESS + "password/change")(
            Method.POST,
        ),
        event,
    );
}

async function change_password_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PasswordChangeResponse, RequestError>>,
    event: FormData,
): Promise<PasswordChangeResult> {
    const current_password = event.get("current_password");
    const new_password = event.get("new_password");

    if (
        typeof current_password !== "string" ||
        typeof new_password !== "string"
    )
        return Err({ status: 400, message: "Missing password fields" });

    const body = new Blob(
        [
            JSON.stringify({
                current_password,
                new_password,
            }),
        ],
        { type: "application/json" },
    );

    const request_result = await handler(body);

    if (!request_result.ok) {
        const status = request_result.error.status_code;

        if (status === 400 || status === 401 || status === 500) {
            try {
                return Err({
                    status,
                    message: (
                        (await request_result.error.response.json()) as {
                            error: string;
                        }
                    ).error,
                });
            } catch {
                return Err({
                    status,
                    message: await request_result.error.response.text(),
                });
            }
        }

        return Err({
            status,
            message: await request_result.error.response.text(),
        });
    }

    return request_result;
}

export const testing_PasswordChanger = change_password_with_handler;
export type testing_PasswordChangeResponse = PasswordChangeResponse;
