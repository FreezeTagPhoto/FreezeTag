/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import {
    formatCamera,
    formatDate,
    formatLocation,
} from "@/common/gallery/format";

describe("common/format", () => {
    it('formatDate returns "—" for null', () => {
        expect(formatDate(null)).toBe("—");
    });

    it("formatDate uses Intl.DateTimeFormat with expected options", () => {
        const fmt: Intl.DateTimeFormat["format"] = jest.fn(() => "FORMATTED");

        const spy = jest
            .spyOn(Intl, "DateTimeFormat")
            .mockImplementation((_locales, _options) => {
                const formatter = {
                    format: fmt,
                } as unknown as Intl.DateTimeFormat;
                return formatter;
            });

        const out = formatDate(1700000000);
        expect(out).toBe("FORMATTED");

        expect(spy).toHaveBeenCalledTimes(1);
        const args = spy.mock.calls[0];

        expect(args[0]).toBe(undefined);
        expect(args[1]).toStrictEqual({
            timeZone: undefined,
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "2-digit",
        });

        expect(fmt).toHaveBeenCalledTimes(1);
        const dateArg = (fmt as jest.Mock).mock.calls[0][0] as Date;
        expect(dateArg).toBeInstanceOf(Date);

        spy.mockRestore();
    });

    it('formatLocation returns "—" when either coordinate is null', () => {
        expect(formatLocation(null, 10)).toBe("—");
        expect(formatLocation(10, null)).toBe("—");
        expect(formatLocation(null, null)).toBe("—");
    });

    it("formatLocation formats 5 decimals", () => {
        expect(formatLocation(40.1234567, -111.9876543)).toBe(
            "40.12346, -111.98765",
        );
    });

    it("formatCamera joins make/model when present and non-empty; otherwise —", () => {
        expect(formatCamera(null, null)).toBe("—");
        expect(formatCamera("  ", "  ")).toBe("—");
        expect(formatCamera("Apple", null)).toBe("Apple");
        expect(formatCamera(null, "iPhone")).toBe("iPhone");
        expect(formatCamera("Apple", "iPhone")).toBe("Apple iPhone");
        expect(formatCamera(" Apple ", " iPhone ")).toBe("Apple iPhone");
    });
});
