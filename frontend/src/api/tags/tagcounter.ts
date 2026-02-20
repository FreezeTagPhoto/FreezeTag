import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { parseUserQuery } from "@/common/search/parse";
import { compileTokensToApiQuery } from "@/common/search/compile";

export type TagCountResult = Result<
    Record<string, number>,
    { status: number; message: string }
>;

type TagCountResponse = Record<string, number>;

export default async function TagCounter(
    search_query: string,
): Promise<TagCountResult> {
    return count_tags_with_handler(
        ApiHandler<TagCountResponse>(SERVER_ADDRESS + "tag/search?")(
            Method.GET,
        ),
        search_query,
    );
}

async function count_tags_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<TagCountResponse, RequestError>>,
    search_query: string,
): Promise<TagCountResult> {
    const tokens = parseUserQuery(search_query);
    const api_query = compileTokensToApiQuery(tokens);
    const result = await handler(api_query);

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
