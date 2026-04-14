const fixedDateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});

/**
 * Formats ISO timestamp as dd.MM.yyyy HH:mm (for example 14.01.2026 20:52).
 */
export function formatDateTime(iso: string): string {
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) {
    return iso;
  }
  return fixedDateTimeFormatter.format(t).replace(',', '');
}
