import Grid from '@mui/material/Unstable_Grid2';
import { Children } from 'react';
import type { ReactNode } from 'react';

export interface TileGridProps {
  children: ReactNode;
}

export function TileGrid({ children }: TileGridProps) {
  return (
    <Grid container spacing={2}>
      {Children.map(children, (child, index) => (
        <Grid key={index} xs={12} sm={6} md={4} lg={3}>
          {child}
        </Grid>
      ))}
    </Grid>
  );
}
