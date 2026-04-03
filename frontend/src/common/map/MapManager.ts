export const MAP_ENABLED_STORAGE_KEY = "freezetag-map-enabled";
export const MAP_CHANGED_EVENT = "freezetag:map-changed";

export const MapEnabledGetter = (): boolean => {
    if (typeof window === "undefined") return true;
    const stored = localStorage.getItem(MAP_ENABLED_STORAGE_KEY);
    // Default to enabled; only explicitly stored "false" disables it
    return stored !== "false";
};

export const MapEnabledSetter = (enabled: boolean): void => {
    localStorage.setItem(MAP_ENABLED_STORAGE_KEY, enabled ? "true" : "false");
};
