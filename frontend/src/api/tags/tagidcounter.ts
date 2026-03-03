import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type TagCountResult = Result<
    Record<string, number>,
    { status: number; message: string }
>;

type TagCountResponse = Record<string, number>;

export default async function TagIdCounter(
    image_ids: number[],
): Promise<TagCountResult> {
    return count_tags_with_handler(
        ApiHandler<TagCountResponse>(SERVER_ADDRESS + "tag/count?")(Method.GET),
        image_ids,
    );
}

async function count_tags_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<TagCountResponse, RequestError>>,
    image_ids: number[],
): Promise<TagCountResult> {
    const result = await handler(image_ids.map((val) => `id=${val}`).join("&"));

    if (!result.ok) {
        const status = result.error.status_code;
        if (status == 400) {
            return Err({
                status,
                message: (
                    (await result.error.response.json()) as { error: string }
                ).error,
            });
        } else {
            return Err({
                status,
                message: await result.error.response.text(),
            });
        }
    }

    return result;
}

export const testing_TagCounter = count_tags_with_handler;
export type testing_TagCountResponse = TagCountResponse;
