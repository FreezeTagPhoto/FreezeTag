"use server";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

/**
 * The string array is a list of errors on an otherwise successful request
 */
export type TagAddResult = Result<
    string[],
    { status: number; message: string }
>;
type TagAddResponse = {
    added: string[];
    errors: string[];
};

export default async function TagAdder(
    image_ids: number[],
    tags: string[],
): Promise<TagAddResult> {
    return add_tag_with_handler(
        ApiHandler<TagAddResponse>(
            SERVER_ADDRESS + "tag/add?",
            false,
        )(Method.POST),
        image_ids,
        tags,
    );
}

async function add_tag_with_handler(
    handler: (data: BodyInit) => Promise<Result<TagAddResponse, RequestError>>,
    image_ids: number[],
    tags: string[],
): Promise<TagAddResult> {
    const query_arr = [];

    for (const image_id of image_ids) query_arr.push(`id=${image_id}`);
    for (const tag of tags) query_arr.push(`tag=${tag}`);
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

    return Ok(result.value.errors);
}

export const testing_TagAdder = add_tag_with_handler;
export type testing_TagAddResponse = TagAddResponse;
