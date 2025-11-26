import { Option, Some, None } from "./option";
type Ok<T> = { ok: true; value: T };
type Err<E> = { ok: false; error: E };
export type Result<T, E> = Ok<T> | Err<E>;
export const Ok = <T>(value: T): Ok<T> => ({ ok: true, value });
export const Err = <E>(error: E): Err<E> => ({ ok: false, error });

export function expect<T, E>(result: Result<T, E>, error_message: string): T {
    if (result.ok) {
        return result.value;
    }
    throw new Error(error_message);
}

export function unwrap<T, E>(result: Result<T, E>): T {
    return expect(result, "Bad unwrap of result!");
}

export function unwrap_or<T, E>(result: Result<T, E>, alt: T): T {
    return unwrap_or_else(result, () => alt);
}

export function unwrap_or_else<T, E>(result: Result<T, E>, alt: () => T): T {
    if (result.ok) {
        return result.value;
    }
    return alt();
}

export function expect_err<T, E>(
    result: Result<T, E>,
    error_message: string,
): E {
    if (!result.ok) {
        return result.error;
    }
    throw new Error(error_message);
}

export function unwrap_err<T, E>(result: Result<T, E>): E {
    return expect_err(result, "Bad unwrap_err of result!");
}

export function flatten<T, E>(result: Result<Result<T, E>, E>): Result<T, E> {
    if (result.ok) {
        return result.value;
    }
    return result;
}

export function transpose<T, E>(
    result: Result<Option<T>, E>,
): Option<Result<T, E>> {
    if (result.ok) {
        if (result.value.some) {
            return Some(Ok(result.value.value));
        } else {
            return None();
        }
    } else {
        return Some(Err(result.error));
    }
}
