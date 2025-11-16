import Alert from '@mui/material/Alert';
import type { ToastMessage } from './types';

export interface ToastProps {
  toast: ToastMessage;
  onClose?: (id: string) => void;
}

export function Toast({ toast, onClose }: ToastProps) {
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
