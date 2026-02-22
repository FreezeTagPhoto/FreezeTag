import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { Plugin } from "@/api/plugins/pluginshelpers";

export type PluginListResult = Result<
    Plugin[],
    { status: number; message: string }
>;
type PluginListResponse = Plugin[];

export default async function PluginsLister(): Promise<PluginListResult> {
    return list_plugins_with_handler(
        ApiHandler<PluginListResponse>(SERVER_ADDRESS + "plugins")(Method.GET),
    );
}

async function list_plugins_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginListResponse, RequestError>>,
): Promise<PluginListResult> {
    const request_result = await handler("");

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
