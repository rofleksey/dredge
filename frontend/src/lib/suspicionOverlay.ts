export type SuspicionOverlayRow = {
  isSus: boolean;
  susDescription?: string;
  susType?: string;
};

/** Prefer live WebSocket overlay when present, else API snapshot. */
export function effectiveChatterIsSus(
  chatterId: number | null | undefined,
  fromApi: boolean,
  overlay: Readonly<Record<number, SuspicionOverlayRow>>,
): boolean {
  if (chatterId == null || !Number.isFinite(chatterId)) {
    return fromApi;
  }
  const row = overlay[chatterId];
  if (row !== undefined) {
    return row.isSus;
  }
  return fromApi;
}

/** Tooltip text from overlay when user is suspicious. */
export function effectiveSuspicionTitle(
  chatterId: number | null | undefined,
  isSusEffective: boolean,
  overlay: Readonly<Record<number, SuspicionOverlayRow>>,
): string | undefined {
  if (!isSusEffective) {
    return undefined;
  }
  if (chatterId == null || !Number.isFinite(chatterId)) {
    return undefined;
  }
  const d = overlay[chatterId]?.susDescription?.trim();
  return d || undefined;
}
