import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PluginGetConfigResult = Result<
    PluginGetConfigResponse,
    { status: number; message: string }
>;

export type PluginConfigField = {
    value: string;
    default?: string;
    protected: boolean;
    name?: string;
    description?: string;
};

export type PluginGetConfigResponse = Record<string, PluginConfigField>;

export type PluginSetConfigResult = Result<
    PluginSetConfigResponse,
    { status: number; message: string }
>;

export type PluginSetConfigResponse = string;

export async function GetPluginConfig(
    plugin: string,
): Promise<PluginGetConfigResult> {
    return get_plugin_config_with_handler(
        ApiHandler<PluginGetConfigResponse>(
            SERVER_ADDRESS + "/plugins/config",
            false,
        )(Method.GET),
        plugin,
    );
}

export async function SetPluginConfig(
    plugin: string,
    changes: Record<string, string>,
): Promise<PluginSetConfigResult> {
    return set_plugin_config_with_handler(
        ApiHandler<PluginSetConfigResponse>(
            SERVER_ADDRESS +
                `/plugins/config?plugin=${encodeURIComponent(plugin)}`,
            true,
        )(Method.POST),
        changes,
    );
}

async function get_plugin_config_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginGetConfigResponse, RequestError>>,
    plugin: string,
): Promise<PluginGetConfigResult> {
    const request_result = await handler(`?plugin=${plugin}`);
    if (!request_result.ok) {
        return Err({
            status: request_result.error.status_code,
            message: (
                (await request_result.error.response.json()) as {
                    error: string;
                }
            ).error,
        });
    }
    return request_result;
}

async function set_plugin_config_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginSetConfigResponse, RequestError>>,
    changes: Record<string, string>,
): Promise<PluginSetConfigResult> {
    const request_result = await handler(`${JSON.stringify(changes)}`);
    if (!request_result.ok) {
        return Err({
            status: request_result.error.status_code,
            message: (
                (await request_result.error.response.json()) as {
                    error: string;
                }
            ).error,
        });
    }
    return request_result;
}
