import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type FormDataPluginRunnerResult = Result<
    FormDataPluginRunnerResponse,
    { status: number; message: string }
>;

export type FormDataPluginRunnerResponse = string;

// Pass in a single ID as a number if you want to use a single_image signature hook, otherwise use the array for image_batch
export default async function FormDataPluginRunner(
    plugin: string,
    hook: string,
    form_data: FormData,
): Promise<FormDataPluginRunnerResult> {
    return run_plugin_with_handler(
        ApiHandler<FormDataPluginRunnerResponse>(
            SERVER_ADDRESS +
                `plugins/run?plugin=${encodeURIComponent(plugin)}&hook=${encodeURIComponent(hook)}`,
            true,
        )(Method.POST),
        form_data,
    );
}

async function run_plugin_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<FormDataPluginRunnerResponse, RequestError>>,
    form_data: FormData,
): Promise<FormDataPluginRunnerResult> {
    // Source for parsing code: https://stackoverflow.com/a/46774073
    const object: Record<string, FormDataEntryValue | FormDataEntryValue[]> =
        {};
    form_data.forEach((value, key) => {
        // Reflect.has in favor of: object.hasOwnProperty(key)
        if (!Reflect.has(object, key)) {
            object[key] = value;
            return;
        }
        if (!Array.isArray(object[key])) {
            object[key] = [object[key]];
        }
        object[key].push(value);
    });
    const json = JSON.stringify(object);

    const request_result = await handler(json);

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
