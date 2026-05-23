export function toNumberValue(value: unknown, fallback = 0) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }

  return fallback;
}

export function toOptionalNumber(value: unknown) {
  if (value === undefined || value === null || value === '') {
    return undefined;
  }

  return toNumberValue(value);
}

export function isNumericValue(value: unknown, expected: number) {
  return toOptionalNumber(value) === expected;
}

export function isZeroStatus(value: unknown) {
  return isNumericValue(value, 0);
}
