import Box from '@mui/material/Box';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import type { PanelDescriptor } from './types';
import { ScrollableViewPort } from './ScrollableViewPort';
import { DraggableDivider } from './DraggableDivider';
import { useState } from 'react';

export interface ConcertinaShellProps {
  panels: PanelDescriptor[];
}

export function ConcertinaShell({ panels }: ConcertinaShellProps) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [leftWidth, setLeftWidth] = useState(50);
  if (!panels.length) {
    return null;
  }

  if (isMobile || panels.length === 1) {
    const panel = panels[panels.length - 1];
    return (
      <Box
        className="concertina-shell concertina-shell-single"
        sx={{
          height: '100%',
          display: 'flex',
          flex: 1,
          minHeight: 0,
          alignItems: 'stretch',
          p: 0,
          boxSizing: 'border-box',
          backgroundColor: (theme) => theme.palette.background.paper,
        }}
      >
        <ScrollableViewPort>{panel.content}</ScrollableViewPort>
      </Box>
    );
  }

  const leftPanel = panels[panels.length - 2];
  const rightPanel = panels[panels.length - 1];

  return (
      <Box
        className="concertina-shell"
        display="flex"
        height="100%"
        flex={1}
        sx={{
          overflow: 'hidden',
          minHeight: 0,
          alignItems: 'stretch',
          p: 0,
          boxSizing: 'border-box',
          backgroundColor: (theme) => theme.palette.background.paper,
        }}
      >
      <ScrollableViewPort
        sx={{
          flex: `0 0 ${leftWidth}%`,
          minWidth: 200,
        }}
      >
        {leftPanel.content}
      </ScrollableViewPort>
      <DraggableDivider
        onResize={(delta) => {
          setLeftWidth((prev) => {
            const containerWidth = window.innerWidth;
            const percentDelta = (delta / containerWidth) * 100;
            const next = Math.min(80, Math.max(20, prev + percentDelta));
            return next;
          });
        }}
      />
      <ScrollableViewPort
        sx={{
          flex: '1 1 auto',
          minWidth: 200,
        }}
      >
        {rightPanel.content}
      </ScrollableViewPort>
    </Box>
  );
}
