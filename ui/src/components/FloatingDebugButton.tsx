import BugReportIcon from '@mui/icons-material/BugReport';
import { useCallback, useEffect, useRef, useState } from 'react';

interface Position {
  x: number;
  y: number;
}

export interface FloatingDebugButtonProps {
  message?: string;
  onClick?: () => void;
  initialPosition?: Position;
}

const BUTTON_SIZE = 48;
const CLICK_THRESHOLD = 5;

export function FloatingDebugButton({
  message = 'Debug clicked',
  onClick,
  initialPosition = { x: 24, y: 24 },
}: FloatingDebugButtonProps) {
  const [position, setPosition] = useState<Position>(initialPosition);
  const [isDragging, setIsDragging] = useState(false);
  const posRef = useRef(position);
  const dragInfoRef = useRef({
    offsetX: 0,
    offsetY: 0,
    startX: 0,
    startY: 0,
    moved: false,
    dragging: false,
  });

  useEffect(() => {
    posRef.current = position;
  }, [position]);

  const clampPosition = useCallback((x: number, y: number): Position => {
    const maxX = window.innerWidth - BUTTON_SIZE;
    const maxY = window.innerHeight - BUTTON_SIZE;
    return {
      x: Math.min(Math.max(0, x), Math.max(0, maxX)),
      y: Math.min(Math.max(0, y), Math.max(0, maxY)),
    };
  }, []);

  const updatePosition = useCallback(
    (clientX: number, clientY: number) => {
      const info = dragInfoRef.current;
      const nextPos = clampPosition(clientX - info.offsetX, clientY - info.offsetY);
      setPosition(nextPos);
    },
    [clampPosition],
  );

  const handleRelease = useCallback(
    (shouldTriggerClick: boolean) => {
      dragInfoRef.current.dragging = false;
      dragInfoRef.current.moved = false;
      setIsDragging(false);
      if (shouldTriggerClick) {
        dumpPanelDiagnostics(message);
        onClick?.();
      }
    },
    [message, onClick],
  );

  useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (event: MouseEvent) => {
      if (!dragInfoRef.current.dragging) return;
      event.preventDefault();
      const { startX, startY, moved } = dragInfoRef.current;
      if (!moved) {
        const distance = Math.hypot(event.clientX - startX, event.clientY - startY);
        if (distance > CLICK_THRESHOLD) {
          dragInfoRef.current.moved = true;
        }
      }
      updatePosition(event.clientX, event.clientY);
    };

    const handleMouseUp = (event: MouseEvent) => {
      if (!dragInfoRef.current.dragging) return;
      event.preventDefault();
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
      handleRelease(!dragInfoRef.current.moved);
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [handleRelease, isDragging, updatePosition]);

  useEffect(() => {
    if (!isDragging) return;

    const handleTouchMove = (event: TouchEvent) => {
      if (!dragInfoRef.current.dragging) return;
      const touch = event.touches[0];
      if (!touch) return;
      event.preventDefault();
      const { startX, startY, moved } = dragInfoRef.current;
      if (!moved) {
        const distance = Math.hypot(touch.clientX - startX, touch.clientY - startY);
        if (distance > CLICK_THRESHOLD) {
          dragInfoRef.current.moved = true;
        }
      }
      updatePosition(touch.clientX, touch.clientY);
    };

    const handleTouchEnd = (event: TouchEvent) => {
      if (!dragInfoRef.current.dragging) return;
      event.preventDefault();
      window.removeEventListener('touchmove', handleTouchMove);
      window.removeEventListener('touchend', handleTouchEnd);
      handleRelease(!dragInfoRef.current.moved);
    };

    window.addEventListener('touchmove', handleTouchMove, { passive: false });
    window.addEventListener('touchend', handleTouchEnd);

    return () => {
      window.removeEventListener('touchmove', handleTouchMove);
      window.removeEventListener('touchend', handleTouchEnd);
    };
  }, [handleRelease, isDragging, updatePosition]);

  const beginDrag = (clientX: number, clientY: number) => {
    const current = posRef.current;
    dragInfoRef.current = {
      offsetX: clientX - current.x,
      offsetY: clientY - current.y,
      startX: clientX,
      startY: clientY,
      moved: false,
      dragging: true,
    };
    setIsDragging(true);
  };

  const handleMouseDown = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    beginDrag(event.clientX, event.clientY);
  };

  const handleTouchStart = (event: React.TouchEvent<HTMLButtonElement>) => {
    const touch = event.touches[0];
    if (!touch) return;
    event.preventDefault();
    beginDrag(touch.clientX, touch.clientY);
  };

  const style: React.CSSProperties = {
    position: 'fixed',
    top: position.y,
    left: position.x,
    width: BUTTON_SIZE,
    height: BUTTON_SIZE,
    borderRadius: '50%',
    border: 'none',
    backgroundColor: '#d32f2f',
    color: '#fff',
    boxShadow: '0 4px 10px rgba(0,0,0,0.25)',
    cursor: isDragging ? 'grabbing' : 'grab',
    zIndex: 9999,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 18,
    userSelect: 'none',
  };

  return (
    <button
      type="button"
      aria-label="Floating debug button"
      style={style}
      onMouseDown={handleMouseDown}
      onTouchStart={handleTouchStart}
    >
      <BugReportIcon fontSize="small" />
    </button>
  );
}

function dumpPanelDiagnostics(customMessage: string) {
  const appMainMetrics = measureElement(document.querySelector('.app-main'));
  const concertinaMetrics = measureElement(document.querySelector('.concertina-shell'));
  const panelReport = Array.from(document.querySelectorAll('.panel-container-debug'))
    .map((panel, index) => {
      const metrics = measureElement(panel);
      if (!metrics) return null;
      const scroller = panel.querySelector('.panel-content-scroller');
      const scrollerMetrics = scroller ? measureElement(scroller) : null;
      return {
        panel: index,
        width: metrics.width,
        height: metrics.height,
        scrollHeight: metrics.scrollHeight,
        clientHeight: metrics.clientHeight,
        scrollTop: metrics.scrollTop,
        scrollerScrollHeight: scrollerMetrics?.scrollHeight ?? null,
        scrollerClientHeight: scrollerMetrics?.clientHeight ?? null,
      };
    })
    .filter(isNotNull);

  console.group(`FloatingDebugButton: ${customMessage}`);
  console.log('Viewport', { innerWidth: window.innerWidth, innerHeight: window.innerHeight });
  console.log('App main', appMainMetrics ?? 'not found');
  console.log('Concertina shell', concertinaMetrics ?? 'not found');
  if (panelReport.length) {
    console.log('Panel container sizes');
    console.table(panelReport);
  }
  console.groupEnd();
}

interface ElementMetrics {
  width: number;
  height: number;
  top: number;
  bottom: number;
  left: number;
  right: number;
  scrollHeight: number;
  clientHeight: number;
  scrollTop: number;
}

function measureElement(element: Element | null): ElementMetrics | null {
  if (!element || !(element instanceof HTMLElement)) {
    return null;
  }
  const rect = element.getBoundingClientRect();
  return {
    width: Math.round(rect.width),
    height: Math.round(rect.height),
    top: Math.round(rect.top),
    bottom: Math.round(rect.bottom),
    left: Math.round(rect.left),
    right: Math.round(rect.right),
    scrollHeight: element.scrollHeight,
    clientHeight: element.clientHeight,
    scrollTop: element.scrollTop,
  };
}

function isNotNull<T>(value: T | null): value is T {
  return value !== null;
}
