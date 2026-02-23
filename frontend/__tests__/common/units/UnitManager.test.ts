import {
    ApplyUnits,
    UNIT_STORAGE_KEY,
    UnitsGetter,
    UnitsNearDefault,
    UnitsSetter,
    type UnitSystem,
} from "@/common/units/UnitManager";

class LocalStorageMock {
    private store = new Map<string, string>();

    getItem(key: string): string | null {
        return this.store.has(key) ? this.store.get(key)! : null;
    }
    setItem(key: string, value: string) {
        this.store.set(key, String(value));
    }
    removeItem(key: string) {
        this.store.delete(key);
    }
    clear() {
        this.store.clear();
    }
}

function setNavigatorLanguage(value: string | undefined) {
    Object.defineProperty(globalThis.navigator, "language", {
        value,
        configurable: true,
    });
}

describe("UnitManager (jsdom)", () => {
    let ls: LocalStorageMock;

    beforeEach(() => {
        ls = new LocalStorageMock();

        Object.defineProperty(globalThis, "localStorage", {
            value: ls as unknown as Storage,
            configurable: true,
        });

        setNavigatorLanguage("en-GB");
        ls.clear();
    });

    describe("UnitsGetter", () => {
        test("returns stored metric", () => {
            ls.setItem(UNIT_STORAGE_KEY, "metric");
            expect(UnitsGetter()).toBe("metric");
        });

        test("returns stored imperial", () => {
            ls.setItem(UNIT_STORAGE_KEY, "imperial");
            expect(UnitsGetter()).toBe("imperial");
        });

        test("ignores invalid stored value and falls back to locale region", () => {
            ls.setItem(UNIT_STORAGE_KEY, "nope");
            setNavigatorLanguage("en-US");
            expect(UnitsGetter()).toBe("imperial");
        });

        test("returns imperial for US/LR/MM regions", () => {
            ls.setItem(UNIT_STORAGE_KEY, "invalid");

            setNavigatorLanguage("en-US");
            expect(UnitsGetter()).toBe("imperial");

            setNavigatorLanguage("en-LR");
            expect(UnitsGetter()).toBe("imperial");

            setNavigatorLanguage("en-MM");
            expect(UnitsGetter()).toBe("imperial");
        });

        test("returns metric for non-imperial regions", () => {
            ls.setItem(UNIT_STORAGE_KEY, "invalid");

            setNavigatorLanguage("en-GB");
            expect(UnitsGetter()).toBe("metric");

            setNavigatorLanguage("fr-FR");
            expect(UnitsGetter()).toBe("metric");
        });

        test("returns metric when navigator.language is missing/empty", () => {
            ls.setItem(UNIT_STORAGE_KEY, "invalid");

            setNavigatorLanguage(undefined);
            expect(UnitsGetter()).toBe("metric");

            setNavigatorLanguage("");
            expect(UnitsGetter()).toBe("metric");
        });
    });

    describe("UnitsSetter", () => {
        test("stores the unit system in localStorage", () => {
            UnitsSetter("imperial");
            expect(ls.getItem(UNIT_STORAGE_KEY)).toBe("imperial");
        });

        test("accepts both metric and imperial", () => {
            const values: UnitSystem[] = ["metric", "imperial"];
            for (const v of values) {
                UnitsSetter(v);
                expect(ls.getItem(UNIT_STORAGE_KEY)).toBe(v);
            }
        });
    });

    describe("UnitsNearDefault", () => {
        test("returns km for metric", () => {
            ls.setItem(UNIT_STORAGE_KEY, "metric");
            expect(UnitsNearDefault()).toBe("km");
        });

        test("returns mi for imperial", () => {
            ls.setItem(UNIT_STORAGE_KEY, "imperial");
            expect(UnitsNearDefault()).toBe("mi");
        });

        test("defaults to metric => km when nothing is stored and locale is non-imperial", () => {
            setNavigatorLanguage("en-GB");
            expect(UnitsNearDefault()).toBe("km");
        });

        test("defaults to imperial => mi when nothing is stored and locale is US", () => {
            setNavigatorLanguage("en-US");
            expect(UnitsNearDefault()).toBe("mi");
        });
    });

    describe("ApplyUnits", () => {
        test("sets data-units on documentElement", () => {
            ApplyUnits("imperial");
            expect(document.documentElement.getAttribute("data-units")).toBe(
                "imperial",
            );

            ApplyUnits("metric");
            expect(document.documentElement.getAttribute("data-units")).toBe(
                "metric",
            );
        });
    });
});
