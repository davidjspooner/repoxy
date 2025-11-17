import Box from '@mui/material/Box';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import type { PanelDescriptor } from './types';
import { PanelContainer } from './PanelContainer';
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
          border: '3px solid #2e7d32',
          p: 0,
          boxSizing: 'border-box',
          backgroundColor: (theme) => theme.palette.background.paper,
        }}
      >
        <PanelContainer>{panel.content}</PanelContainer>
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
        border: '3px solid #2e7d32',
        p: 0,
        boxSizing: 'border-box',
        backgroundColor: (theme) => theme.palette.background.paper,
      }}
    >
      <PanelContainer
        className="panel-slot panel-slot-left"
        decorated={false}
        sx={{
          flex: `0 0 ${leftWidth}%`,
          minWidth: 200,
        }}
      >
        {leftPanel.content}
      </PanelContainer>
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
      <PanelContainer
        className="panel-slot panel-slot-right"
        decorated={false}
        sx={{
          flex: '1 1 auto',
          minWidth: 200,
        }}
      >
        {rightPanel.content}
      </PanelContainer>
    </Box>
  );
}
