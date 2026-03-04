"use client";

import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Ok } from "@/common/result";

export type FileDeleteResponse = {
    id: number;
    file: string;
};

export default function FileDeleter(id: number) {
    return async (): Promise<Result<FileDeleteResponse, RequestError>> => {
        return delete_file_with_handler(
            ApiHandler<FileDeleteResponse>(SERVER_ADDRESS + "file/delete/")(
                Method.DELETE,
            ),
            id,
        );
    };
}

async function delete_file_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<FileDeleteResponse, RequestError>>,
    id: number,
): Promise<Result<FileDeleteResponse, RequestError>> {
    const result = await handler(String(id));

    if (!result.ok) return result;

    return Ok(result.value);
}

export const testing_FileDeleter = delete_file_with_handler;
export type testing_FileDeleteResponse = FileDeleteResponse;
