import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumNamesListResult = Result<
    string[],
    { status: number; message: string }
>;

type AlbumNamesListResponse = string[];

export default async function AlbumNamesLister(): Promise<AlbumNamesListResult> {
    return list_album_names_with_handler(
        ApiHandler<AlbumNamesListResponse>(SERVER_ADDRESS + "album/names")(
            Method.GET,
        ),
    );
}

async function list_album_names_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumNamesListResponse, RequestError>>,
): Promise<AlbumNamesListResult> {
    const result = await handler("");

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        if (status === 400) {
            let message = bodyText;
            try {
                message = (JSON.parse(bodyText) as { error?: string }).error ||
                    bodyText;
            } catch {
                // keep raw response text if backend returned plain text
            }
            return Err({
                status,
                message,
            });
        }
        return Err({
            status,
            message: bodyText,
        });
    }

    return result;
}

export const testing_AlbumNamesLister = list_album_names_with_handler;
