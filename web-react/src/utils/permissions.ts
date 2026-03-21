export function hasPermission(
  permissions: string[],
  expected?: string | string[],
): boolean {
  if (!expected) {
    return true;
  }

  const required = Array.isArray(expected) ? expected : [expected];

  return required.some((permission) => {
    if (permissions.includes('*') || permissions.includes(permission)) {
      return true;
    }

    return permissions.some((candidate) => {
      if (!candidate.endsWith('.*')) {
        return false;
      }

      const prefix = candidate.slice(0, -2);
      return permission.startsWith(`${prefix}.`);
    });
  });
}
