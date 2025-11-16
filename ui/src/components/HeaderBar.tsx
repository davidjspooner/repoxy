import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import AccountCircle from '@mui/icons-material/AccountCircle';
import { useState } from 'react';

export interface HeaderBarProps {
  title?: string;
  onOpenSettings?: () => void;
}

export function HeaderBar({ title = 'Repoxy', onOpenSettings }: HeaderBarProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  return (
    <AppBar position="fixed">
      <Toolbar>
        <Typography variant="h6" sx={{ flexGrow: 1 }}>
          {title}
        </Typography>
        <IconButton size="large" color="inherit" onClick={(event) => setAnchorEl(event.currentTarget)}>
          <AccountCircle />
        </IconButton>
        <Menu anchorEl={anchorEl} open={open} onClose={() => setAnchorEl(null)}>
          <MenuItem onClick={() => {
            onOpenSettings?.();
            setAnchorEl(null);
          }}>Settings…</MenuItem>
          <MenuItem disabled>Account…</MenuItem>
          <MenuItem disabled>Logout</MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  );
}
