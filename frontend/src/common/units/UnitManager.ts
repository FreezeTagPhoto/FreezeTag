export type UnitSystem = "metric" | "imperial";

export const UNIT_STORAGE_KEY = "freezetag-unit-option";

export const UnitsGetter: () => UnitSystem = () => {
    if (typeof window === "undefined") return "metric";

    const stored = localStorage.getItem(UNIT_STORAGE_KEY);
    if (stored === "metric" || stored === "imperial") return stored;

    const lang = (navigator.language ?? "").toLowerCase();
    const region = lang.split("-")[1] ?? "";

    // usa, liberia, and myanmar commonly use imperial
    if (region === "us" || region === "lr" || region === "mm")
        return "imperial";

    return "metric";
};

export const UnitsSetter = (units: UnitSystem) => {
    localStorage.setItem(UNIT_STORAGE_KEY, units);
};

export const UnitsNearDefault = (): "km" | "mi" => {
    return UnitsGetter() === "imperial" ? "mi" : "km";
};

export const ApplyUnits = (units: UnitSystem) => {
    if (typeof document === "undefined") return;
    document.documentElement.setAttribute("data-units", units);
};
