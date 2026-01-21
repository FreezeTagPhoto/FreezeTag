import type { FieldKey } from "./keys";

export type TokenKey = FieldKey | "tag";

export type Token =
    | {
          kind: "field";
          key: Exclude<TokenKey, "tag">;
          valueRaw: string;
          value: string;
          exact: boolean;
          range: { start: number; end: number };
          error?: string;
      }
    | {
          kind: "tag";
          valueRaw: string;
          value: string;
          exact: boolean;
          range: { start: number; end: number };
          error?: string;
      };