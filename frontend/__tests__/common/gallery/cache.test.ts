/**
 * @jest-environment jsdom
 */

import "@testing-library/jest-dom";
import { act, renderHook } from "@testing-library/react";

import { useCachedById } from "@/common/gallery/cache";
import type { Result } from "@/common/result";
import { None, Some } from "@/common/option";

type Err = { message: string };

function ok<T>(value: T): Result<T, Err> {
    return { ok: true, value } as Result<T, Err>;
}
function bad<T>(message: string): Result<T, Err> {
    return { ok: false, error: { message } } as Result<T, Err>;
}

describe("common/gallery/cache.useCachedById", () => {
    it("does nothing when selectedId is null", () => {
        const fetcher = jest.fn(async (_id: number) => ok("x"));

        const { result } = renderHook(() =>
            useCachedById<string, Err>(null, fetcher),
        );

        expect(fetcher).not.toHaveBeenCalled();
        expect(result.current.loading).toBe(false);
        expect(result.current.error).toStrictEqual(None());
        expect(result.current.byId).toStrictEqual({});
        expect(result.current.current).toStrictEqual(None());
    });

    it("fetches when id not cached, then caches and sets current + clears error", async () => {
        const fetcher = jest.fn(async (id: number) => ok(`v:${id}`));

        const { result, rerender } = renderHook(
            ({ selectedId }) => useCachedById<string, Err>(selectedId, fetcher),
            { initialProps: { selectedId: 1 } },
        );

        expect(fetcher).toHaveBeenCalledTimes(1);
        expect(fetcher).toHaveBeenCalledWith(1);

        await act(async () => {});

        expect(result.current.loading).toBe(false);
        expect(result.current.error).toStrictEqual(None());
        expect(result.current.byId).toStrictEqual({ 1: "v:1" });
        expect(result.current.current).toStrictEqual(Some("v:1"));

        // same id again => no new fetch (cached)
        rerender({ selectedId: 1 });
        await act(async () => {});
        expect(fetcher).toHaveBeenCalledTimes(1);

        // new id => fetch again
        rerender({ selectedId: 2 });
        expect(fetcher).toHaveBeenCalledTimes(2);
        expect(fetcher).toHaveBeenLastCalledWith(2);

        await act(async () => {});

        expect(result.current.byId).toStrictEqual({ 1: "v:1", 2: "v:2" });
        expect(result.current.current).toStrictEqual(Some("v:2"));
    });

    it("sets error when fetch fails and does not cache", async () => {
        const fetcher = jest.fn(async (_id: number) => bad<string>("nope"));

        const { result } = renderHook(() =>
            useCachedById<string, Err>(7, fetcher),
        );

        await act(async () => {});

        expect(result.current.loading).toBe(false);
        expect(result.current.error).toStrictEqual(Some("nope"));
        expect(result.current.byId).toStrictEqual({});
        expect(result.current.current).toStrictEqual(None());
    });

    it("does not set state after unmount (cancelled)", async () => {
        let resolve!: (v: Result<string, Err>) => void;

        const fetcher = jest.fn(
            () =>
                new Promise<Result<string, Err>>((r) => {
                    resolve = r;
                }),
        );

        const { unmount } = renderHook(() =>
            useCachedById<string, Err>(10, fetcher),
        );

        expect(fetcher).toHaveBeenCalledTimes(1);

        unmount();

        // resolving after unmount should not throw
        await act(async () => {
            resolve(ok("late"));
        });
    });
});
