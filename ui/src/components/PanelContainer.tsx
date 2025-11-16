import Box from '@mui/material/Box';
import type { ReactNode } from 'react';

export interface PanelContainerProps {
  children: ReactNode;
}

export function PanelContainer({ children }: PanelContainerProps) {
  return (
    <Box
      sx={{
        height: '100%',
        width: '100%',
        overflowY: 'auto',
        overflowX: 'hidden',
        padding: 2,
      }}
    >
      {children}
    </Box>
  );
}
