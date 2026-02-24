// src/api/tags/tagdeleter.ts
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

export type TagDeleteResult = Result<
    number,
    { status: number; message: string }
>;

type TagDeleteResponse = {
    deleted: number;
};

export default async function TagDeleter(
    tags: string[],
): Promise<TagDeleteResult> {
    return delete_tag_with_handler(
        ApiHandler<TagDeleteResponse>(SERVER_ADDRESS + "tag/delete?")(
            Method.DELETE,
        ),
        tags,
    );
}

async function delete_tag_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<TagDeleteResponse, RequestError>>,
    tags: string[],
): Promise<TagDeleteResult> {
    const query_arr: string[] = [];
    for (const tag of tags) query_arr.push(`tag=${encodeURIComponent(tag)}`);

    const result = await handler(query_arr.join("&"));

    if (!result.ok) {
        const status = result.error.status_code;

        const text = await result.error.response.text();
        try {
            const body = JSON.parse(text) as { error?: unknown };
            if (typeof body?.error === "string") {
                return Err({ status, message: body.error });
            }
        } catch {
            // ignore JSON parse errors (for now?)
        }

        return Err({ status, message: text });
    }

    return Ok(result.value.deleted);
}

export const testing_TagDeleter = delete_tag_with_handler;
export type testing_TagDeleteResponse = TagDeleteResponse;
