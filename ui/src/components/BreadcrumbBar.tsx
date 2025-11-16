import Breadcrumbs from '@mui/material/Breadcrumbs';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import type { BreadcrumbItem } from './types';

export interface BreadcrumbBarProps {
  items: BreadcrumbItem[];
}

export function BreadcrumbBar({ items }: BreadcrumbBarProps) {
  if (!items.length) {
    return null;
  }

  return (
    <Breadcrumbs
      separator={<NavigateNextIcon fontSize="small" />}
      aria-label="breadcrumb"
      sx={{ padding: 1 }}
    >
      {items.map((item) =>
        item.isCurrent ? (
          <Typography key={item.id} color="text.primary">
            {item.label}
          </Typography>
        ) : (
          <Link
            key={item.id}
            underline="hover"
            color="inherit"
            component="button"
            onClick={item.onSelect}
          >
            {item.label}
          </Link>
        ),
      )}
    </Breadcrumbs>
  );
}
