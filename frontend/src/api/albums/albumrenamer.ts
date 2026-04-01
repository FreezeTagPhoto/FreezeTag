import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumRenameResult = Result<
    { message: string },
    { status: number; message: string }
>;

type AlbumRenameResponse = { message: string };

export default async function AlbumRenamer(
    oldName: string,
    newName: string,
): Promise<AlbumRenameResult> {
    return rename_album_with_handler(
        ApiHandler<AlbumRenameResponse>(SERVER_ADDRESS + "album/rename")(
            Method.POST,
        ),
        oldName,
        newName,
    );
}

async function rename_album_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumRenameResponse, RequestError>>,
    oldName: string,
    newName: string,
): Promise<AlbumRenameResult> {
    const result = await handler(
        JSON.stringify({ old_name: oldName, new_name: newName }),
    );

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

export const testing_AlbumRenamer = rename_album_with_handler;
