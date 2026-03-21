import type { ReactNode } from 'react';
import { usePermissionStore } from '@/store/permission';

interface PermissionGateProps {
  permission?: string | string[];
  fallback?: ReactNode;
  children: ReactNode;
}

export function PermissionGate({
  permission,
  fallback = null,
  children,
}: PermissionGateProps) {
  const hasPermission = usePermissionStore((state) => state.hasPermission);

  if (!hasPermission(permission)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
