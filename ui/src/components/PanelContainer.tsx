import Box from '@mui/material/Box';
import type { SxProps, Theme } from '@mui/material/styles';
import type { ReactNode } from 'react';

export interface PanelContainerProps {
  children: ReactNode;
  className?: string;
  sx?: SxProps<Theme>;
}

export function PanelContainer({ children, className, sx }: PanelContainerProps) {
  const combinedClassName = ['panel-container-debug', className].filter(Boolean).join(' ');

  return (
    <Box
      className={combinedClassName}
      sx={[
        {
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          width: '100%',
          alignSelf: 'stretch',
          flex: '1 1 auto',
          boxSizing: 'border-box',
          overflow: 'auto',
          backgroundColor: '#f3e5f5',
        },
        sx,
      ]}
    >
      <Box
        sx={{
          height: '100%',
          width: '100%',
          overflow: 'auto',
          padding: 0,
          backgroundColor: 'transparent',
          minHeight: 0,
          boxSizing: 'border-box',
          flex: '1 1 auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {children}
      </Box>
    </Box>
  );
}
