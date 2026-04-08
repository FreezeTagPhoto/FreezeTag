import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result } from "@/common/result";
import { None, Option, Some } from "@/common/option";

export type PluginUploadErr = Option<{ status: number; message: string }>;

type PluginUploadResponse = string;

export default async function PluginUploader(
    link: string,
): Promise<PluginUploadErr> {
    return upload_plugins_with_handler(
        ApiHandler<PluginUploadResponse>(
            SERVER_ADDRESS + "plugins/upload",
            true,
        )(Method.POST),
        link,
    );
}

async function upload_plugins_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<PluginUploadResponse, RequestError>>,
    link: string,
): Promise<PluginUploadErr> {
    const request_result = await handler(`"${link}"`);

    if (!request_result.ok) {
        const status = request_result.error.status_code;
        if (status == 400)
            return Some({
                status,
                message: (
                    (await request_result.error.response.json()) as {
                        error: string;
                    }
                ).error,
            });
        else
            return Some({
                status,
                message: await request_result.error.response.text(),
            });
    }

    return None();
}
