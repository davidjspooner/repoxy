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
          height: '100%',
          maxHeight: '100%',
          alignSelf: 'stretch',
          flex: '1 1 auto',
          boxSizing: 'border-box',
          overflow: 'auto',
          backgroundColor: '#f3e5f5',
          p: (theme) => theme.spacing(2),
        },
        sx,
      ]}
    >
      <Box
        className="panel-content-scroller"
        sx={{
          height: '100%',
          width: '100%',
          overflow: 'auto',
          minHeight: 0,
          boxSizing: 'border-box',
          flex: '1 1 auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <Box
          sx={{
            backgroundColor: '#e3f2fd',
            padding: (theme) => theme.spacing(2),
            borderRadius: 1,
            boxSizing: 'border-box',
            display: 'inline-flex',
            flexDirection: 'column',
            minWidth: 'max-content',
            minHeight: 'max-content',
            gap: (theme) => theme.spacing(1.5),
          }}
        >
          {children}
        </Box>
      </Box>
    </Box>
  );
}
