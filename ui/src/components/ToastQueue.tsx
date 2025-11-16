import Box from '@mui/material/Box';
import Slide from '@mui/material/Slide';
import { useMemo } from 'react';
import type { ToastMessage } from './types';
import { Toast } from './Toast';

export interface ToastQueueProps {
  toasts: ToastMessage[];
  onDismiss?: (id: string) => void;
}

export function ToastQueue({ toasts, onDismiss }: ToastQueueProps) {
  const sortedToasts = useMemo(() => [...toasts], [toasts]);

  return (
    <Box
      sx={{
        position: 'fixed',
        bottom: 16,
        left: '50%',
        transform: 'translateX(-50%)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        zIndex: 1400,
      }}
    >
      {sortedToasts.map((toast) => (
        <Slide key={toast.id} direction="up" in={true} mountOnEnter unmountOnExit>
          <Box>
            <Toast toast={toast} onClose={onDismiss} />
          </Box>
        </Slide>
      ))}
    </Box>
  );
}
