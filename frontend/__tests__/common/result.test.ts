/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import {
    Ok,
    Err,
    unwrap,
    unwrap_err,
    Result,
    unwrap_or,
    unwrap_or_else,
    flatten,
    transpose,
} from "@/common/result";
import { Some, None } from "@/common/option";

describe("Result", () => {
    it("throws on bad unwrap", () => {
        const result = Err(19);
        expect(() => unwrap(result)).toThrow("unwrap");
    });

    it("works on good unwrap", () => {
        const result = Ok(19);
        expect(unwrap(result)).toBe(19);
    });

    it("throws on bad unwrap_err", () => {
        const result = Ok(19);
        expect(() => unwrap_err(result)).toThrow("unwrap");
    });

    it("Works on good unwrap_err", () => {
        const result = Err(19);
        expect(unwrap_err(result)).toBe(19);
    });

    it("has typeguards work for Oks", () => {
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

    it("has typeguards work for Errs", () => {
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

    it("works on unwrap_or with err", () => {
        const result = Err(19);

        expect(result.ok).toBeFalsy();

        const unwrapped = unwrap_or(result, 2);

        expect(unwrapped).toBe(2);
    });

    it("works on unwrap_or with ok", () => {
        const result = Ok(19);

        expect(result.ok).toBeTruthy();

        const unwrapped = unwrap_or(result, 2);

        expect(unwrapped).toBe(19);
    });

    it("works when else throws on unwrap_or_else", () => {
        const result = Err(19);

        expect(() =>
            unwrap_or_else(result, () => {
                throw new Error("Error!");
            }),
        ).toThrow();
    });

    it("works when using flatten", () => {
        const result = Ok(Err(19));
        expect(flatten(result)).toStrictEqual(Err(19));
    });

    it("works when using flatten with an err", () => {
        const result = Err(19);
        expect(flatten(result)).toStrictEqual(Err(19));
    });

    it("works when using transpose", () => {
        const result = Ok(Some(19));
        expect(transpose(result)).toStrictEqual(Some(Ok(19)));
    });

    it("works when using transpose with none", () => {
        const result = Ok(None());
        expect(transpose(result)).toStrictEqual(None());
    });

    it("works when using transpose with an err", () => {
        const result = Err(19);
        expect(transpose(result)).toStrictEqual(Some(Err(19)));
    });
});
