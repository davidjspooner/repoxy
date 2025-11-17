import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import AccountCircle from '@mui/icons-material/AccountCircle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { PropsWithChildren, useState } from 'react';
import type { BreadcrumbItem } from './types';
import { BreadcrumbBar } from './BreadcrumbBar';

export interface HeaderBarProps {
  breadcrumbs: BreadcrumbItem[];
  onOpenSettings?: () => void;
}

function HeaderRightToolbar({ children }: PropsWithChildren) {
  return (
    <Box display="flex" alignItems="center" gap={1.5}>
      {children}
    </Box>
  );
}

export function HeaderBar({ breadcrumbs, onOpenSettings }: HeaderBarProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  return (
    <AppBar
      position="fixed"
      color="default"
      elevation={1}
      sx={{
        backgroundColor: (theme) => theme.palette.background.paper,
        color: (theme) => theme.palette.text.primary,
        borderBottom: (theme) =>
          theme.palette.mode === 'dark' ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(0,0,0,0.1)',
      }}
    >
      <Toolbar sx={{ gap: 2 }}>
        <Box flexGrow={1}>
          <BreadcrumbBar items={breadcrumbs} />
        </Box>
        <HeaderRightToolbar>
          <Typography variant="body2" color="text.secondary">
            Admin
          </Typography>
          <IconButton size="large" color="inherit" onClick={(event) => setAnchorEl(event.currentTarget)}>
            <AccountCircle />
          </IconButton>
        </HeaderRightToolbar>
        <Menu anchorEl={anchorEl} open={open} onClose={() => setAnchorEl(null)}>
          <MenuItem
            onClick={() => {
              onOpenSettings?.();
              setAnchorEl(null);
            }}
          >
            Settings…
          </MenuItem>
          <MenuItem disabled>Account…</MenuItem>
          <MenuItem disabled>Logout</MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  );
}
