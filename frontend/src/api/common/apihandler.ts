"use client";
import { Result, Ok, Err } from "@/common/result";

export const enum Method {
    GET = "GET",
    POST = "POST",
    DELETE = "DELETE",
}

export type RequestError = { status_code: number; response: Response };

/**
 * "json":  response.json()  (default, T should be a plain object type)
 * "blob":  response.blob()  (use when T = Blob, e.g. image downloads)
 */
export function ApiHandler<T>(
    address: string,
    body_request: boolean = true,
    responseType: "json" | "blob" = "json",
) {
    const parseResponse = async (response: Response): Promise<T> => {
        if (responseType === "blob") {
            return (await response.blob()) as T;
        }
        return (await response.json()) as T;
    };

    return (method: Method) => {
        if (method === Method.POST && body_request) {
            return async (data: BodyInit): Promise<Result<T, RequestError>> => {
                let response;
                try {
                    response = await fetch(address, {
                        method: method,
                        body: data,
                    });
                } catch (error) {
                    return Err({
                        status_code: 0,
                        response: new Response(`Catastrophic Error: ${error}`),
                    });
                }

                if (response.ok) {
                    try {
                        return Ok(await parseResponse(response));
                    } catch (error) {
                        return Err({
                            status_code: 0,
                            response: new Response(`Catastrophic! ${error}`),
                        });
                    }
                } else {
                    return Err({ status_code: response.status, response });
                }
            };
        } else {
            return async (data: BodyInit): Promise<Result<T, RequestError>> => {
                let response;
                try {
                    response = await fetch(address + data, {
                        method: method,
                    });
                } catch (error) {
                    return Err({
                        status_code: 0,
                        response: new Response(`Catastrophic Error: ${error}`),
                    });
                }

                if (response.ok) {
                    try {
                        return Ok(await parseResponse(response));
                    } catch (error) {
                        return Err({
                            status_code: 0,
                            response: new Response(`Catastrophic! ${error}`),
                        });
                    }
                } else {
                    return Err({ status_code: response.status, response });
                }
            };
        }
    };
}
