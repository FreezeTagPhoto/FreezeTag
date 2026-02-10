// src/hooks/useCachedById.ts
import { useEffect, useMemo, useState } from "react";
import type { Result } from "@/common/result";
import { None, Some, type Option } from "@/common/option";

type WithMessage = { message: string };

/**
 * Caches fetched values by numeric id. Designed for patterns like:
 * - selectedId changes
 * - if cached, reuse
 * - else fetch -> set loading/error -> cache
 */
export function useCachedById<T, E extends WithMessage = WithMessage>(
  selectedId: number | null,
  fetcher: (id: number) => Promise<Result<T, E>>,
) {
  const [byId, setById] = useState<Record<number, T>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Option<string>>(None());

  const current: Option<T> = useMemo(() => {
    if (selectedId === null) return None();
    const v = byId[selectedId];
    return v === undefined ? None() : Some(v);
  }, [selectedId, byId]);

  useEffect(() => {
    if (selectedId === null) return;

    // already cached
    if (byId[selectedId] !== undefined) {
      setError(None());
      setLoading(false);
      return;
    }

    let cancelled = false;

    (async () => {
      setLoading(true);
      setError(None());

      const res = await fetcher(selectedId);
      if (cancelled) return;

      if (!res.ok) {
        setError(Some(res.error.message));
        setLoading(false);
        return;
      }

      setById((prev) => ({ ...prev, [selectedId]: res.value }));
      setLoading(false);
    })();

    return () => {
      cancelled = true;
    };
  }, [selectedId, byId, fetcher]);

  return {
    byId,
    setById,
    current,
    loading,
    error,
    setError,
    setLoading,
  };
}