import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type TagGetResult = Result<
    string[],
    { status: number; message: string }
>;
type TagGetResponse = string[];

/**
 *
 * @param image_id Undefined means that this query will get all tags. Otherwise it just gets the tags for this image
 * @returns A TagGetResult promise
 */
export default async function TagGetter(
    image_id?: number,
): Promise<TagGetResult> {
    return get_tag_with_handler(
        ApiHandler<TagGetResponse>((await SERVER_ADDRESS()) + "tag/list")(
            Method.GET,
        ),
        image_id,
    );
}

async function get_tag_with_handler(
    handler: (data: BodyInit) => Promise<Result<TagGetResponse, RequestError>>,
    image_id?: number,
): Promise<TagGetResult> {
    const result = await (image_id ? handler(`/${image_id}`) : handler(""));

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

export const testing_TagGetter = get_tag_with_handler;
export type testing_TagGetResponse = TagGetResponse;
