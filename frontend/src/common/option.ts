import { Result, Ok, Err } from "./result";

type Some<T> = { some: true; value: T };
type None = { some: false };
export type Option<T> = Some<T> | None;
export const Some = <T>(value: T): Some<T> => ({ some: true, value });
export const None = (): None => ({ some: false });

export function expect<T>(option: Option<T>, error_message: string): T {
  if (option.some) {
    return option.value;
  }
  throw new Error(error_message);
}

export function unwrap<T>(option: Option<T>): T {
  return expect(option, "Bad unwrap of option!");
}

export function unwrap_or<T>(option: Option<T>, alt: T): T {
  return unwrap_or_else(option, () => alt);
}

export function unwrap_or_else<T>(option: Option<T>, alt: () => T): T {
  if (option.some) {
    return option.value;
  }
  return alt();
}

export function ok_or<T, E>(option: Option<T>, err: E): Result<T, E> {
  return ok_or_else(option, () => err);
}

export function ok_or_else<T, E>(
  option: Option<T>,
  err: () => E,
): Result<T, E> {
  if (option.some) {
    return Ok(option.value);
  }
  return Err(err());
}

export function flatten<T>(option: Option<Option<T>>): Option<T> {
  if (option.some) {
    return option.value;
  }
  return option;
}

export function transpose<T, E>(
  option: Option<Result<T, E>>,
): Result<Option<T>, E> {
  if (option.some) {
    if (option.value.ok) {
      return Ok(Some(option.value.value));
    } else {
      return Err(option.value.error);
    }
  }
  return Ok(None());
}
