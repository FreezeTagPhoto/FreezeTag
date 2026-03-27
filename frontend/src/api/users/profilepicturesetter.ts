import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

type ProfilePictureSetResponse = { message: string };

export type ProfilePictureSetResult = Result<
    ProfilePictureSetResponse,
    { status: number; message: string }
>;

/**
 * POST /users/profile-picture/{id}
 * Uploads a new profile picture for the given user (multipart form, field: "picture").
 * A user can only update their own profile picture.
 */
export default async function ProfilePictureSetter(
    user_id: number,
    picture: File,
): Promise<ProfilePictureSetResult> {
    const fd = new FormData();
    fd.set("picture", picture);

    const handler = ApiHandler<ProfilePictureSetResponse>(
        SERVER_ADDRESS + `users/profile-picture/${user_id}`,
    )(Method.POST);

    const request_result = await handler(fd);

    if (!request_result.ok) {
        const status = request_result.error.status_code;
        try {
            return Err({
                status,
                message: (
                    (await request_result.error.response.json()) as {
                        error: string;
                    }
                ).error,
            });
        } catch {
            return Err({
                status,
                message:
                    (await request_result.error.response
                        .text()
                        .catch(() => "")) || "Failed to update profile picture",
            });
        }
    }

    return request_result;
}

export const testing_ProfilePictureSetter = ProfilePictureSetter;
export type testing_ProfilePictureSetResponse = ProfilePictureSetResponse;
