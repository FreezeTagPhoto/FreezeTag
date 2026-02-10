export function normalizeTag(s: string) {
  return s.trim().replace(/\s+/g, " ");
}

function isSubsequence(needle: string, hay: string) {
  let i = 0;
  for (let j = 0; j < hay.length && i < needle.length; j++) {
    if (hay[j] === needle[i]) i++;
  }
  return i === needle.length;
}

export function rankTag(tag: string, needleRaw: string) {
  const needle = needleRaw.toLowerCase();
  const t = tag.toLowerCase();

  if (t === needle) return 0;
  if (t.startsWith(needle)) return 1;

  const idx = t.indexOf(needle);
  if (idx !== -1) return 2 + idx / 100;

  if (isSubsequence(needle, t)) return 3;

  return 999;
}

export function computeTagSuggestions(args: {
  addOpen: boolean;
  allTags: string[] | null;
  currentTags: string[] | null;
  addValue: string;
  tagSuggestPinned: boolean;
  limit?: number;
}) {
  const { addOpen, allTags, currentTags, addValue, tagSuggestPinned, limit = 10 } = args;

  if (!addOpen) return [];
  if (!allTags || allTags.length === 0) return [];

  const current = new Set((currentTags ?? []).map((t) => t));
  const candidates = allTags.filter((t) => !current.has(t));

  const needle = normalizeTag(addValue);
  const allowEmpty = tagSuggestPinned;

  if (!needle) {
    if (!allowEmpty) return [];
    return [...candidates].sort((a, b) => a.localeCompare(b)).slice(0, limit);
  }

  return candidates
    .map((t) => ({ tag: t, score: rankTag(t, needle) }))
    .filter((x) => x.score < 999)
    .sort((a, b) => a.score - b.score || a.tag.localeCompare(b.tag))
    .slice(0, limit)
    .map((x) => x.tag);
}