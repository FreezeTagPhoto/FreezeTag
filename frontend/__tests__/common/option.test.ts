/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import {
  Some,
  None,
  unwrap,
  Option,
  unwrap_or,
  unwrap_or_else,
  flatten,
  transpose,
  ok_or,
} from "@/common/option";

import { Ok, Err } from "@/common/result";

describe("Option", () => {
  it("throws on bad unwrap", () => {
    const result = None();
    expect(() => unwrap(result)).toThrow("unwrap");
  });

  it("works on good unwrap", () => {
    const result = Some(19);
    expect(unwrap(result)).toBe(19);
  });

  it("has typeguards work for Oks", () => {
    function sample(): Option<number> {
      return Some(19);
    }

    const result = sample();
    if (result.some) {
      expect(result.value).toBe(19);
    } else {
      fail();
    }

    const result2 = sample();
    switch (result2.some) {
      case true:
        expect(result2.value).toBe(19);
        break;
      case false:
        fail();
    }
  });

  it("has typeguards work for Errs", () => {
    function sample(): Option<number> {
      return None();
    }

    const result = sample();
    if (result.some) {
      fail();
    } else {
      expect(true).toBeTruthy();
    }
  });

  it("works on unwrap_or with err", () => {
    const result = None();

    expect(result.some).toBeFalsy();

    const unwrapped = unwrap_or(result, 2);

    expect(unwrapped).toBe(2);
  });

  it("works on unwrap_or with ok", () => {
    const result = Some(19);

    expect(result.some).toBeTruthy();

    const unwrapped = unwrap_or(result, 2);

    expect(unwrapped).toBe(19);
  });

  it("works when else throws on unwrap_or_else", () => {
    const result = None();

    expect(() =>
      unwrap_or_else(result, () => {
        throw new Error("Error!");
      }),
    ).toThrow();
  });

  it("works when using ok_or on none", () => {
    const result = None();
    expect(ok_or(result, 5)).toStrictEqual(Err(5));
  });

  it("works when using ok_or on some", () => {
    const result = Some(19);
    expect(ok_or(result, 5)).toStrictEqual(Ok(19));
  });

  it("works when using flatten", () => {
    const result = Some(None());
    expect(flatten(result)).toStrictEqual(None());
  });

  it("works when using flatten with an err", () => {
    const result = None();
    expect(flatten(result)).toStrictEqual(None());
  });

  it("works when using transpose", () => {
    const result = Some(Ok(19));
    expect(transpose(result)).toStrictEqual(Ok(Some(19)));
  });

  it("works when using transpose with none", () => {
    const result = None();
    expect(transpose(result)).toStrictEqual(Ok(None()));
  });

  it("works when using transpose with an err", () => {
    const result = Some(Err(19));
    expect(transpose(result)).toStrictEqual(Err(19));
  });
});
