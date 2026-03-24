import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type PluginRunnerResult = Result<
    PluginRunnerResponse,
    { status: number; message: string }
>;

export type PluginRunnerResponse = string;

// Pass in a single ID as a number if you want to use a single_image signature hook, otherwise use the array for image_batch
export default async function PluginRunner(
    plugin: string,
    hook: string,
    image_ids: number | number[],
): Promise<PluginRunnerResult> {
    return run_plugin_with_handler(
        ApiHandler<PluginRunnerResponse>(
            SERVER_ADDRESS +
                `plugins/run?plugin=${encodeURIComponent(plugin)}&hook=${encodeURIComponent(hook)}`,
            true,
        )(Method.POST),
        image_ids,
    );
}

async function run_plugin_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginRunnerResponse, RequestError>>,
    image_ids: number | number[],
): Promise<PluginRunnerResult> {
    let request_result;
    if (typeof image_ids === "number") {
        request_result = await handler(`${image_ids}`);
    } else {
        request_result = await handler(`${JSON.stringify(image_ids)}`);
    }

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
