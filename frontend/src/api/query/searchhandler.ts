"use server";
import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type SearchResult = Result<
  number[],
  { status: number; message: string }
>;
type SearchResponse = number[];

const regex = /\s*(.*?(?:".*?")?)\s*(?:,|$)/g;

export default async function SearchHandler(
  user_query: string,
): Promise<SearchResult> {
  return search_with_handler(
    ApiHandler<SearchResponse>(SERVER_ADDRESS + "search/?")(Method.GET),
    user_query,
  );
}

async function search_with_handler(
  handler: (data: BodyInit) => Promise<Result<SearchResponse, RequestError>>,
  user_query: string,
): Promise<SearchResult> {
  const queries = user_query.matchAll(regex);
  const compiled_api_queries = [];

  for (const query of queries) {
    const query_string = query[1];
    let new_query = "";

    if (query_string === "") {
      continue;
    } else if (query_string.startsWith("make=")) {
      const sub_query = query_string.slice("make=".length);
      if (sub_query.startsWith(`"`)) {
        new_query = "make=" + sub_query.slice(1, sub_query.length - 1);
      } else {
        new_query = "makeLike=" + sub_query;
      }
    } else if (query_string.startsWith("model=")) {
      const sub_query = query_string.slice("model=".length);
      if (sub_query.startsWith(`"`)) {
        new_query = "model=" + sub_query.slice(1, sub_query.length - 1);
      } else {
        new_query = "modelLike=" + sub_query;
      }
    } else if (
      query_string.startsWith("near=") ||
      query_string.startsWith("takenBefore=") ||
      query_string.startsWith("takenAfter=") ||
      query_string.startsWith("uploadedBefore=") ||
      query_string.startsWith("uploadedAfter=")
    ) {
      new_query = query_string;
    } else {
      // Handle tags
      if (query_string.startsWith(`"`)) {
        new_query = "tag=" + query_string.slice(1, query_string.length - 1);
      } else {
        new_query = "tagLike=" + query_string;
      }
    }

    compiled_api_queries.push(new_query);
  }
  const api_query = compiled_api_queries.join("&");

  const result = await handler(api_query);

  if (!result.ok) {
    const status = result.error.status_code;
    if (status == 400) {
      return Err({
        status,
        message: ((await result.error.response.json()) as { error: string })
          .error,
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
