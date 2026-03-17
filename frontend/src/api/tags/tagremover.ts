import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

/**
 * The string array is a list of errors on an otherwise successful request
 */
export type TagRemoveResult = Result<
    string[],
    { status: number; message: string }
>;
type TagRemoveResponse = {
    deleted: { count: number; id: number }[];
    errors: { id: number; reason: string }[];
};

export default async function TagRemover(
    image_ids: number[],
    tags: string[],
): Promise<TagRemoveResult> {
    return remove_tag_with_handler(
        ApiHandler<TagRemoveResponse>(SERVER_ADDRESS + "tag/remove?")(
            Method.DELETE,
        ),
        image_ids,
        tags,
    );
}

async function remove_tag_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<TagRemoveResponse, RequestError>>,
    image_ids: number[],
    tags: string[],
): Promise<TagRemoveResult> {
    const query_arr = [];

    for (const image_id of image_ids) query_arr.push(`id=${image_id}`);
    for (const tag of tags) query_arr.push(`tag=${encodeURIComponent(tag)}`);
    const result = await handler(query_arr.join("&"));

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

    return Ok(
        result.value.errors.map((err) => `Image Id ${err.id}: ${err.reason}`),
    );
}

export const testing_TagRemover = remove_tag_with_handler;
export type testing_TagRemoveResponse = TagRemoveResponse;
