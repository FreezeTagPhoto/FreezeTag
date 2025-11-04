/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import { Ok, Err, unwrap, unwrap_err, Result } from "@/common/result";

describe("Result", () => {
  it("Throws on bad unwrap", () => {
    const result = Err(19);
    expect(() => unwrap(result)).toThrow("unwrap");
  });

  it("Works on good unwrap", () => {
    const result = Ok(19);
    expect(unwrap(result)).toBe(19);
  });

  it("Throws on bad unwrap_err", () => {
    const result = Ok(19);
    expect(() => unwrap_err(result)).toThrow("unwrap");
  });

  it("Works on good unwrap_err", () => {
    const result = Err(19);
    expect(unwrap_err(result)).toBe(19);
  });

  it("Typeguards work for Oks", () => {
    function sample(): Result<number, number> {
      return Ok(19);
    }

    const result = sample();
    if (result.ok) {
      expect(result.value).toBe(19);
    } else {
      fail();
    }

    const result2 = sample();
    switch (result2.ok) {
      case true:
        expect(result2.value).toBe(19);
        break;
      case false:
        fail();
    }
  });

  it("Typeguards work for Errs", () => {
    function sample(): Result<number, number> {
      return Err(19);
    }

    const result = sample();
    if (result.ok) {
      fail();
    } else {
      expect(result.error).toBe(19);
    }
  });
});
