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
      <PanelContainer>
        {panel.content}
      </PanelContainer>
    );
  }

  const leftPanel = panels[panels.length - 2];
  const rightPanel = panels[panels.length - 1];

  return (
    <Box display="flex" height="100%">
      <Box flex={`0 0 ${leftWidth}%`} minWidth={200} borderRight="1px solid #1f2a36">
        <PanelContainer>{leftPanel.content}</PanelContainer>
      </Box>
      <DraggableDivider onResize={(delta) => {
        setLeftWidth((prev) => {
          const containerWidth = window.innerWidth;
          const percentDelta = (delta / containerWidth) * 100;
          const next = Math.min(80, Math.max(20, prev + percentDelta));
          return next;
        });
      }} />
      <Box flex="1" minWidth={200}>
        <PanelContainer>{rightPanel.content}</PanelContainer>
      </Box>
    </Box>
  );
}
