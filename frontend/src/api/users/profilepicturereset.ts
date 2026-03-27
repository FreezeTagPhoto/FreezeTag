import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

type ProfilePictureResetResponse = { message: string };

export type ProfilePictureResetResult = Result<
    ProfilePictureResetResponse,
    { status: number; message: string }
>;

/**
 * DELETE /users/profile-picture/{id}
 * Resets the profile picture for the given user to the default generated avatar.
 * A user can only reset their own profile picture.
 */
export default async function ProfilePictureReset(
    user_id: number,
): Promise<ProfilePictureResetResult> {
    const handler = ApiHandler<ProfilePictureResetResponse>(
        SERVER_ADDRESS + `users/profile-picture/${user_id}`,
        false,
    )(Method.DELETE);

    const request_result = await handler("");

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
                        .catch(() => "")) || "Failed to reset profile picture",
            });
        }
    }

    return request_result;
}

export const testing_ProfilePictureReset = ProfilePictureReset;
export type testing_ProfilePictureResetResponse = ProfilePictureResetResponse;
