import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PluginAbleResult = Result<
    PluginAbleResponse,
    { status: number; message: string }
>;
type PluginAbleResponse = { disabled: boolean };

export default async function PluginsAbler(
    plugin: string,
    enabled: boolean,
): Promise<PluginAbleResult> {
    return able_plugins_with_handler(
        ApiHandler<PluginAbleResponse>(
            SERVER_ADDRESS + "plugins",
            false,
        )(Method.POST),
        plugin,
        enabled,
    );
}

async function able_plugins_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginAbleResponse, RequestError>>,
    plugin: string,
    enabled: boolean,
): Promise<PluginAbleResult> {
    const request_result = await handler(
        `?plugin=${plugin}&enabled=${enabled ? "true" : "false"}`,
    );

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
