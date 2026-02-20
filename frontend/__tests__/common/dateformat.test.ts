import { formatLongDate, formatShortDate } from "@/common/dateformat";

describe("Date Format", () => {
    it("Handles short date correctly", () => {
        const result = formatShortDate(1771553884);
        expect(result).toBe("Feb 19, 2026");
    });

    it("Handles short date correctly with milliseconds", () => {
        const result = formatShortDate(1771553884000);
        expect(result).toBe("Feb 19, 2026");
    });

    it("Handles short date null correctly", () => {
        const result = formatShortDate(null);
        expect(result).toBe("—");
    });

    it("Handles short date correctly with timezone", () => {
        const result = formatShortDate(1771553884, { timeZone: "UTC" });
        expect(result).toBe("Feb 20, 2026");
    });

    it("Handles long date correctly", () => {
        const result = formatLongDate(1771553884);
        expect(result).toBe("Feb 19, 2026, 7:18 PM");
    });

    it("Handles long date correctly with milliseconds", () => {
        const result = formatLongDate(1771553884000);
        expect(result).toBe("Feb 19, 2026, 7:18 PM");
    });

    it("Handles long date null correctly", () => {
        const result = formatLongDate(null);
        expect(result).toBe("—");
    });

    it("Handles long date correctly with timezone", () => {
        const result = formatLongDate(1771553884, { timeZone: "UTC" });
        expect(result).toBe("Feb 20, 2026, 2:18 AM");
    });
});
