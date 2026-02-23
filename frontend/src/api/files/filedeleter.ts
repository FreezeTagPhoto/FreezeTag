"use client";

import SERVER_ADDRESS from "@/api/common/serveraddress";
import { type RequestError } from "@/api/common/apihandler";
import { type Result, Ok, Err } from "@/common/result";

export type FileDeleteResponse = {
    id: number;
    file: string;
};

export default function FileDeleter(id: number) {
    return async (): Promise<Result<FileDeleteResponse, RequestError>> => {
        let response: Response;

        try {
            response = await fetch(`${SERVER_ADDRESS}/file/delete/${id}`, {
                method: "DELETE",
                headers: {
                    accept: "application/json",
                },
            });
        } catch (error) {
            return Err({
                status_code: 0,
                response: new Response(`Catastrophic Error: ${error}`),
            });
        }

        if (!response.ok) {
            return Err({ status_code: response.status, response });
        }

        try {
            return Ok((await response.json()) as FileDeleteResponse);
        } catch (error) {
            return Err({
                status_code: response.status,
                response: new Response(`Invalid JSON response: ${error}`),
            });
        }
    };
}
