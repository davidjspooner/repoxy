import Box from '@mui/material/Box';
import type { SxProps, Theme } from '@mui/material/styles';
import type { ReactNode } from 'react';

export interface ScrollableViewPortProps {
  children: ReactNode;
  className?: string;
  sx?: SxProps<Theme>;
}

export function ScrollableViewPort({ children, className, sx }: ScrollableViewPortProps) {
  const combinedClassName = ['ScrollableViewPort', className].filter(Boolean).join(' ');

  return (
    <Box
      className={combinedClassName}
      sx={[
        {
          display: 'block',
          flex: '1 1 auto',
          width: '100%',
          maxWidth: '100%',
          height: '100%',
          maxHeight: '100%',
          minHeight: 0,
          minWidth: 0,
          boxSizing: 'border-box',
          overflowX: 'auto',
          overflowY: 'auto',
          backgroundColor: (theme) => theme.palette.background.paper,
          p: (theme) => theme.spacing(2),
          WebkitOverflowScrolling: 'touch',
        },
        sx,
      ]}
    >
      <Box
        className="ScrollableViewPort-content"
        sx={{
          display: 'inline-flex',
          flexDirection: 'column',
          minWidth: '100%',
          boxSizing: 'border-box',
          flexShrink: 0,
        }}
      >
        {children}
      </Box>
    </Box>
  );
}
