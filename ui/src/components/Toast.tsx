import Alert from '@mui/material/Alert';
import { useEffect } from 'react';
import type { ToastMessage } from './types';

export interface ToastProps {
  toast: ToastMessage;
  onClose?: (id: string) => void;
  timeoutMs?: number;
}

export function Toast({ toast, onClose, timeoutMs = 5000 }: ToastProps) {
  useEffect(() => {
    const timeout = setTimeout(() => {
      onClose?.(toast.id);
    }, timeoutMs);
    return () => clearTimeout(timeout);
  }, [onClose, timeoutMs, toast.id]);

  return (
    <Alert
      severity={toast.level === 'info' ? 'info' : toast.level}
      variant="filled"
      onClose={() => onClose?.(toast.id)}
      sx={{ mb: 1, minWidth: 300 }}
    >
      {toast.message}
    </Alert>
  );
}
