import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import { Children } from 'react';
import type { ReactNode } from 'react';

export interface TileGridProps {
  children: ReactNode;
}

export function TileGrid({ children }: TileGridProps) {
  const items = Children.toArray(children);
  return (
    <Stack direction="row" flexWrap="wrap" spacing={2} useFlexGap flexBasis={260} alignItems="stretch">
      {items.map((child, index) => (
        <Box key={index} sx={{ flex: '0 0 260px', maxWidth: '100%' }}>
          {child}
        </Box>
      ))}
    </Stack>
  );
}
