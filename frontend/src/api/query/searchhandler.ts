"use client";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";
import { parseUserQuery } from "@/common/search/parse";
import { compileTokensToApiQuery } from "@/common/search/compile";

export type SearchResult = Result<
    number[],
    { status: number; message: string }
>;
type SearchResponse = number[];

export default async function SearchHandler(
    user_query: string,
): Promise<SearchResult> {
    return search_with_handler(
        ApiHandler<SearchResponse>((await SERVER_ADDRESS()) + "search?")(
            Method.GET,
        ),
        user_query,
    );
}

async function search_with_handler(
    handler: (data: BodyInit) => Promise<Result<SearchResponse, RequestError>>,
    user_query: string,
): Promise<SearchResult> {
    const tokens = parseUserQuery(user_query);
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

export const testing_SearchHandler = search_with_handler;
export type testing_SearchResponse = SearchResponse;
