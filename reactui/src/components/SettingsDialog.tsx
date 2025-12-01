import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import FormControl from '@mui/material/FormControl';
import FormLabel from '@mui/material/FormLabel';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';

export interface SettingsDialogProps {
  open: boolean;
  themeMode: 'light' | 'dark' | 'system';
  density: 'comfortable' | 'compact';
  apiBaseUrl: string;
  onThemeChange?: (value: SettingsDialogProps['themeMode']) => void;
  onDensityChange?: (value: SettingsDialogProps['density']) => void;
  onClose?: () => void;
}

export function SettingsDialog({
  open,
  themeMode,
  density,
  apiBaseUrl,
  onThemeChange,
  onDensityChange,
  onClose,
}: SettingsDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Settings</DialogTitle>
      <DialogContent dividers>
        <FormControl fullWidth sx={{ mb: 3 }}>
          <FormLabel>Theme</FormLabel>
          <RadioGroup
            value={themeMode}
            onChange={(event) => onThemeChange?.(event.target.value as SettingsDialogProps['themeMode'])}
          >
            <FormControlLabel value="light" control={<Radio />} label="Light" />
            <FormControlLabel value="dark" control={<Radio />} label="Dark" />
            <FormControlLabel value="system" control={<Radio />} label="System" />
          </RadioGroup>
        </FormControl>
        <FormControl fullWidth>
          <FormLabel>Density</FormLabel>
          <RadioGroup
            value={density}
            onChange={(event) => onDensityChange?.(event.target.value as SettingsDialogProps['density'])}
          >
            <FormControlLabel value="comfortable" control={<Radio />} label="Comfortable" />
            <FormControlLabel value="compact" control={<Radio />} label="Compact" />
          </RadioGroup>
        </FormControl>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
