import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';

export interface FooterSummaryBarProps {
  summary: string;
}

export function FooterSummaryBar({ summary }: FooterSummaryBarProps) {
  return (
    <AppBar position="fixed" color="primary" sx={{ top: 'auto', bottom: 0 }}>
      <Toolbar variant="dense">
        <Typography variant="body2" noWrap>
          {summary}
        </Typography>
      </Toolbar>
    </AppBar>
  );
}
