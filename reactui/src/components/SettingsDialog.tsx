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
import FormHelperText from '@mui/material/FormHelperText';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

export interface SettingsDialogProps {
  open: boolean;
  themeMode: 'light' | 'dark' | 'system';
  density: 'comfortable' | 'compact';
  dataSource: 'simulated' | 'backend';
  apiBaseUrl: string;
  onThemeChange?: (value: SettingsDialogProps['themeMode']) => void;
  onDensityChange?: (value: SettingsDialogProps['density']) => void;
  onDataSourceChange?: (value: SettingsDialogProps['dataSource']) => void;
  onClose?: () => void;
}

export function SettingsDialog({
  open,
  themeMode,
  density,
  dataSource,
  apiBaseUrl,
  onThemeChange,
  onDensityChange,
  onDataSourceChange,
  onClose,
}: SettingsDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Settings</DialogTitle>
      <DialogContent dividers>
        <FormControl fullWidth sx={{ mb: 3 }}>
          <FormLabel>Data Source</FormLabel>
          <RadioGroup
            value={dataSource}
            onChange={(event) => onDataSourceChange?.(event.target.value as SettingsDialogProps['dataSource'])}
          >
            <FormControlLabel value="simulated" control={<Radio />} label="Simulated (built-in demo data)" />
            <FormControlLabel
              value="backend"
              control={<Radio />}
              label={
                <Box>
                  <Typography variant="body1">Backend (current origin)</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {apiBaseUrl}
                  </Typography>
                </Box>
              }
            />
          </RadioGroup>
          <FormHelperText>Select whether to use mock data or call the live REST API.</FormHelperText>
        </FormControl>
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
