import { useCallback, useEffect, useRef } from 'react';
import Box from '@mui/material/Box';

export interface DraggableDividerProps {
  onResize?: (delta: number) => void;
}

export function DraggableDivider({ onResize }: DraggableDividerProps) {
  const dragging = useRef(false);
  const lastX = useRef(0);

  const handleMouseDown = useCallback((event: React.MouseEvent<HTMLDivElement>) => {
    dragging.current = true;
    lastX.current = event.clientX;
    event.preventDefault();
  }, []);

  const handleMouseMove = useCallback(
    (event: MouseEvent) => {
      if (!dragging.current) return;
      const delta = event.clientX - lastX.current;
      lastX.current = event.clientX;
      onResize?.(delta);
    },
    [onResize],
  );

  const handleMouseUp = useCallback(() => {
    dragging.current = false;
  }, []);

  // Attach listeners once.
  useEventListener('mousemove', handleMouseMove);
  useEventListener('mouseup', handleMouseUp);

  return (
    <Box
      role="separator"
      tabIndex={0}
      onMouseDown={handleMouseDown}
      sx={{
        width: '6px',
        cursor: 'col-resize',
        backgroundColor: '#1f2a36',
        '&:hover': { backgroundColor: '#2c3947' },
      }}
    />
  );
}

function useEventListener(eventName: string, handler: (event: any) => void) {
  const savedHandler = useRef(handler);

  savedHandler.current = handler;

  useEffect(() => {
    const listener = (event: Event) => savedHandler.current(event);
    window.addEventListener(eventName, listener);
    return () => {
      window.removeEventListener(eventName, listener);
    };
  }, [eventName]);
}
