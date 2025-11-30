import { useState } from 'react';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { SettingsDialog } from '../components';

export function SettingsDialogPanel() {
  const [open, setOpen] = useState(false);
  const [themeMode, setThemeMode] = useState<'light' | 'dark' | 'system'>('light');
  const [density, setDensity] = useState<'comfortable' | 'compact'>('comfortable');
  const [dataSource, setDataSource] = useState<'simulated' | 'backend'>('backend');

  const apiBaseUrl =
    typeof window === 'undefined' ? '/api/ui/v1/' : `${window.location.protocol}//${window.location.host}/api/ui/v1/`;

  return (
    <Stack spacing={2} alignItems="flex-start">
      <Typography variant="body1">
        The Settings dialog is launched from the user menu in the header. Use this panel to preview the modal behaviour in isolation.
      </Typography>
      <Button variant="contained" onClick={() => setOpen(true)}>
        Open Settings
      </Button>
      <SettingsDialog
        open={open}
        onClose={() => setOpen(false)}
        themeMode={themeMode}
        density={density}
        dataSource={dataSource}
        apiBaseUrl={apiBaseUrl}
        onThemeChange={setThemeMode}
        onDensityChange={setDensity}
        onDataSourceChange={setDataSource}
      />
    </Stack>
  );
}
